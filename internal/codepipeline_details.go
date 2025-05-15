package internal

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/codepipeline"
	"github.com/bporter816/aws-tui/internal/repo"
	"github.com/bporter816/aws-tui/internal/ui"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type CodePipelineDetails struct {
	*ui.Tree
	repo   *repo.CodePipeline
	name   string
	app    *Application
}

func NewCodePipelineDetails(repo *repo.CodePipeline, name string, app *Application) *CodePipelineDetails {
	root := tview.NewTreeNode(name)
	root.SetReference("")
	d := &CodePipelineDetails{
		Tree:  ui.NewTree(root),
		repo:  repo,
		name:  name,
		app:   app,
	}
	d.SetSelectedFunc(d.selectHandler)
	return d
}

func (d *CodePipelineDetails) GetLabels() []string {
	return []string{d.name, "Details"}
}

func (d *CodePipelineDetails) GetKeyActions() []KeyAction {
	return nil
}

func (d *CodePipelineDetails) selectHandler(n *tview.TreeNode) {
	// No-op for now, could expand to show action details
}

func (d *CodePipelineDetails) Render() {
	pipeline, err := d.repo.GetPipeline(d.name)
	if err != nil {
		panic(err)
	}
	state, err := d.repo.CpClient.GetPipelineState(
		context.TODO(),
		&codepipeline.GetPipelineStateInput{Name: &d.name},
	)
	if err != nil {
		panic(err)
	}
	// Build a map: stageName -> actionName -> status
	statusMap := map[string]map[string]string{}
	for _, stage := range state.StageStates {
		if stage.StageName == nil { continue }
		if _, ok := statusMap[*stage.StageName]; !ok {
			statusMap[*stage.StageName] = map[string]string{}
		}
		for _, action := range stage.ActionStates {
			if action.ActionName != nil && action.LatestExecution != nil {
				statusMap[*stage.StageName][*action.ActionName] = string(action.LatestExecution.Status)
			}
		}
	}
	root := d.GetRoot()
	root.ClearChildren()
	for _, stage := range pipeline.Pipeline.Stages {
		stageNode := tview.NewTreeNode(*stage.Name)
		stageNode.SetReference(*stage.Name)
		for _, action := range stage.Actions {
			actionLabel := *action.Name + " (" + string(action.ActionTypeId.Category) + ")"
			actionNode := tview.NewTreeNode(actionLabel)
			status := statusMap[*stage.Name][*action.Name]
			switch status {
			case "Succeeded":
				actionNode.SetColor(tcell.ColorGreen) // green
			case "Failed":
				actionNode.SetColor(tcell.ColorRed) // red
			default:
				actionNode.SetColor(tview.Styles.PrimaryTextColor)
			}
			stageNode.AddChild(actionNode)
		}
		root.AddChild(stageNode)
	}
}

func (d *CodePipelineDetails) GetService() string {
	return "CodePipeline"
}
