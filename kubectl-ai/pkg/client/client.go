package client

import (
	"context"
	"fmt"
	"github.com/octoboy233/kubectl-ai/pkg/helper"
	"io"
)

const (
	typeChatGPT = "ChatGPT"
	typeSpark   = "Spark"
)

func CreateCompletion(ctx context.Context, opt *helper.Options, prompt string, writer io.Writer, spinner Spinner) error {
	var cli Client
	switch opt.Typ {
	case typeChatGPT:
		cli = NewChatGPTClient(opt.Token)
	case typeSpark:
		cli = NewSparkClient(opt.AppID, opt.APISecret, opt.APIKey)
	default:
		return fmt.Errorf("invalid type %s", opt.Typ)
	}
	err := cli.CreateCompletion(ctx, prompt, writer, spinner)
	if err != nil {
		return fmt.Errorf("create completion: %w", err)
	}
	return nil
}
