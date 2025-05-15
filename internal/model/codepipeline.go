package model

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/codepipeline/types"
)

type CodePipelineSummary struct {
	Name        *string
	Created     *time.Time
	Updated     *time.Time
	PipelineArn *string
}

func NewCodePipelineSummary(s types.PipelineSummary) CodePipelineSummary {
	return CodePipelineSummary{
		Name:        s.Name,
		Created:     s.Created,
		Updated:     s.Updated,
	}
}
