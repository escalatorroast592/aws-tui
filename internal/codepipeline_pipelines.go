package internal

import (
	"log"
	"os"

	"github.com/bporter816/aws-tui/internal/model"
	"github.com/bporter816/aws-tui/internal/repo"
	"github.com/bporter816/aws-tui/internal/ui"
	"github.com/bporter816/aws-tui/internal/view"
)

type CodePipelinePipelines struct {
	*ui.Table
	view.CodePipeline
	repo *repo.CodePipeline
	app  *Application
	model []model.CodePipelineSummary
}

func NewCodePipelinePipelines(repo *repo.CodePipeline, app *Application) *CodePipelinePipelines {
	return &CodePipelinePipelines{
		Table: ui.NewTable([]string{
			"NAME",
		}, 1, 0),
		repo: repo,
		app:  app,
	}
}

func (c *CodePipelinePipelines) selectHandler(row, col int) {
	if row <= 0 || row > len(c.model) {
		return
	}
	pipeline := c.model[row-1]
	if pipeline.Name == nil {
		return
	}
	detailsView := NewCodePipelineDetails(c.repo, *pipeline.Name, c.app)
	c.app.AddAndSwitch(detailsView)
}

func (c *CodePipelinePipelines) Render() {
	// Set up logging to file
	logFile, err := os.OpenFile("codepipeline_pipelines.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		log.SetOutput(logFile)
		defer logFile.Close()
	} else {
		log.Println("Failed to open log file:", err)
	}

	pipelines, err := c.repo.ListPipelines()
	if (err != nil) {
		panic(err)
	}
	c.model = pipelines

	//add logs what have been received
	logFile.WriteString("Received pipelines:\n")

	var data [][]string
	for _, v := range pipelines {
		data = append(data, []string{
			*v.Name,
		})
	}

	c.SetData(data)
	c.SetSelectedFunc(c.selectHandler)
}

func (c *CodePipelinePipelines) GetLabels() []string     { return []string{"Pipelines"} }
func (c *CodePipelinePipelines) GetKeyActions() []KeyAction { return nil }
