package internal

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/bporter816/aws-tui/internal/repo"
	"github.com/bporter816/aws-tui/internal/ui"
	"github.com/gdamore/tcell/v2"
)

type CodeBuildLogsView struct {
	*ui.Text
	repo         *repo.CodePipeline
	pipelineName string
	stageName    string
	actionName   string
	app          *Application
}

// Accept *repo.CodePipeline directly, not interface{}
func NewCodeBuildLogsView(repo *repo.CodePipeline, pipelineName, stageName, actionName string, app *Application) *CodeBuildLogsView {
	view := &CodeBuildLogsView{
		Text:        ui.NewText(true, "plain"),
		repo:        repo,
		pipelineName: pipelineName,
		stageName:    stageName,
		actionName:   actionName,
		app:         app,
	}
	// Increase scroll speed: PageDown/PageUp scroll 10 lines at a time
	view.Text.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, col := view.Text.GetScrollOffset()
		switch event.Key() {
		case tcell.KeyPgDn:
			view.Text.ScrollTo(row+50, col)
			return nil
		case tcell.KeyPgUp:
			view.Text.ScrollTo(row-50, col)
			return nil
		}
		return event
	})
	view.SetText(fmt.Sprintf("Logs for %s / %s / %s\n(implement log fetching here)", pipelineName, stageName, actionName))
	return view
}

func (v *CodeBuildLogsView) GetLabels() []string {
	return []string{v.pipelineName, v.stageName, v.actionName, "Logs"}
}

func (v *CodeBuildLogsView) GetKeyActions() []KeyAction { return nil }
func (v *CodeBuildLogsView) GetService() string { return "CodeBuildLogs" }

func (v *CodeBuildLogsView) Render() {
	// 1. Get pipeline definition and extract CodeBuild project name
	pipeline, err := v.repo.GetPipeline(v.pipelineName)
	if err != nil {
		v.SetText("Error fetching pipeline: " + err.Error())
		return
	}
	var projectName string
	for _, stage := range pipeline.Pipeline.Stages {
		if stage.Name != nil && *stage.Name == v.stageName {
			for _, action := range stage.Actions {
				if action.Name != nil && *action.Name == v.actionName {
					if action.Configuration != nil {
						if pn, ok := action.Configuration["ProjectName"]; ok {
							projectName = pn
						}
					}
				}
			}
		}
	}
	if projectName == "" {
		v.SetText("Could not determine CodeBuild project name for this action.")
		return
	}
	// 2. List builds for the project
	cbClient := codebuild.NewFromConfig(v.app.GetAWSConfig())
	buildsOut, err := cbClient.ListBuildsForProject(context.TODO(), &codebuild.ListBuildsForProjectInput{
		ProjectName: &projectName,
	})
	if err != nil || len(buildsOut.Ids) == 0 {
		v.SetText("No builds found for project or error: " + err.Error())
		return
	}
	// 3. Get the latest build
	buildOut, err := cbClient.BatchGetBuilds(context.TODO(), &codebuild.BatchGetBuildsInput{Ids: buildsOut.Ids[:1]})
	if err != nil || len(buildOut.Builds) == 0 {
		v.SetText("No build details found or error: " + err.Error())
		return
	}
	build := buildOut.Builds[0]
	if build.Logs == nil || build.Logs.GroupName == nil || build.Logs.StreamName == nil {
		v.SetText("No logs found for build")
		return
	}
	// 4. Fetch logs from CloudWatch Logs
	cwClient := cloudwatchlogs.NewFromConfig(v.app.GetAWSConfig())
	logsOut, err := cwClient.GetLogEvents(context.TODO(), &cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  build.Logs.GroupName,
		LogStreamName: build.Logs.StreamName,
	})
	if err != nil {
		v.SetText("Error fetching logs: " + err.Error())
		return
	}
	var logs string
	for _, event := range logsOut.Events {
		logs += *event.Message
	}
	v.SetText(logs)
}
