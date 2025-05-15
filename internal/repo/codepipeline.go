package repo

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codepipeline"
	"github.com/bporter816/aws-tui/internal/model"
)

type CodePipeline struct {
	cpClient *codepipeline.Client
}

func NewCodePipeline(cpClient *codepipeline.Client) *CodePipeline {
	return &CodePipeline{
		cpClient: cpClient,
	}
}

func (c CodePipeline) ListPipelines() ([]model.CodePipelineSummary, error) {
	pg := codepipeline.NewListPipelinesPaginator(
		c.cpClient,
		&codepipeline.ListPipelinesInput{},
	)
	var pipelines []model.CodePipelineSummary
	for pg.HasMorePages() {
		out, err := pg.NextPage(context.TODO())
		if err != nil {
			return []model.CodePipelineSummary{}, err
		}
		for _, v := range out.Pipelines {
			pipelines = append(pipelines, model.NewCodePipelineSummary(v))
		}
	}
	return pipelines, nil
}

func (c CodePipeline) GetPipeline(name string) (*codepipeline.GetPipelineOutput, error) {
	return c.cpClient.GetPipeline(
		context.TODO(),
		&codepipeline.GetPipelineInput{
			Name: aws.String(name),
		},
	)
}

func (c CodePipeline) ListTags(resourceArn string) (model.Tags, error) {
	out, err := c.cpClient.ListTagsForResource(
		context.TODO(),
		&codepipeline.ListTagsForResourceInput{
			ResourceArn: aws.String(resourceArn),
		},
	)
	if err != nil {
		return model.Tags{}, err
	}
	var tags model.Tags
	for _, v := range out.Tags {
		tags = append(tags, model.Tag{Key: *v.Key, Value: *v.Value})
	}
	return tags, nil
}
