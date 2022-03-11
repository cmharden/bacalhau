package system

import (
	"fmt"
	"os"
	"strings"

	"github.com/filecoin-project/bacalhau/internal/types"
)

type ResultsList struct {
	Node   string
	Cid    string
	Folder string
}

func GetJobData(host string, port int, jobId string) (*types.Job, error) {
	args := &types.ListArgs{}
	result := &types.ListResponse{}
	err := JsonRpcMethod(host, port, "List", args, result)
	if err != nil {
		return nil, err
	}

	for _, jobData := range result.Jobs {
		if strings.HasPrefix(jobData.Id, jobId) {
			return jobData, nil
		}
	}

	return nil, fmt.Errorf("Could not find job: %s", jobId)
}

func GetJobResults(host string, port int, jobId string) (*[]ResultsList, error) {

	job, err := GetJobData(host, port, jobId)

	if err != nil {
		return nil, err
	}

	return ProcessJobIntoResults(job)
}

func ProcessJobIntoResults(job *types.Job) (*[]ResultsList, error) {
	results := []ResultsList{}

	for node := range job.State {

		cid := ""

		if len(job.State[node].Outputs) > 0 {
			cid = job.State[node].Outputs[0].Cid
		}

		results = append(results, ResultsList{
			Node:   node,
			Cid:    cid,
			Folder: GetResultsDirectory(job.Id, node),
		})
	}

	return &results, nil
}

func FetchJobResult(results ResultsList) error {
	resultsFolder, err := GetSystemDirectory(results.Folder)
	if err != nil {
		return err
	}
	if _, err := os.Stat(resultsFolder); !os.IsNotExist(err) {
		return nil
	}
	fmt.Printf("Fetching results for job %s ---> %s\n", results.Cid, results.Folder)
	resultsFolder, err = EnsureSystemDirectory(results.Folder)
	if err != nil {
		return err
	}
	err = RunCommand("ipfs", []string{
		"get",
		results.Cid,
		"--output",
		resultsFolder,
	})
	if err != nil {
		return err
	}
	return nil
}

func FetchJobResults(host string, port int, jobId string) error {
	data, err := GetJobResults(host, port, jobId)
	if err != nil {
		return err
	}

	for _, row := range *data {
		err = FetchJobResult(row)
		if err != nil {
			return err
		}
	}

	return nil
}
