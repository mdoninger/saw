package blade

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"os"
	"time"

	"github.com/TylerBrock/colorjson"
	sawconfig "github.com/TylerBrock/saw/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/fatih/color"
)

// A Blade is a Saw execution instance
type Blade struct {
	config *sawconfig.Configuration
	aws    *sawconfig.AWSConfiguration
	output *sawconfig.OutputConfiguration
	cwl    *cloudwatchlogs.Client
}

// NewBlade creates a new Blade with CloudWatchLogs instance from provided sawconfig
func NewBlade(
	sawconfig *sawconfig.Configuration,
	awsConfig *sawconfig.AWSConfiguration,
	outputConfig *sawconfig.OutputConfiguration,
) *Blade {
	blade := Blade{}
	//awsCfg := aws.Config{}

	//if awsConfig.Region != "" {
	//	awsCfg.Region = awsConfig.Region
	//}
	//
	//if awsConfig.Profile != "" {
	//	awsSessionOpts.Profile = awsConfig.Profile
	//}

	optionsFunc := func(options *stscreds.AssumeRoleOptions) {
		options.TokenProvider = stscreds.StdinTokenProvider
	}
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithAssumeRoleCredentialOptions(optionsFunc))

	if err != nil {
		panic(err)
	}

	blade.cwl = cloudwatchlogs.NewFromConfig(cfg)
	blade.config = sawconfig
	blade.output = outputConfig

	return &blade
}

// GetLogGroups gets the log groups from AWS given the blade configuration
func (b *Blade) GetLogGroups() []types.LogGroup {
	input := b.config.DescribeLogGroupsInput()
	groups := make([]types.LogGroup, 0)
	logGroups, err := b.cwl.DescribeLogGroups(context.TODO(), input)
	if err != nil {
		return groups
	}

	for _, group := range logGroups.LogGroups {
		groups = append(groups, group)
	}
	return groups
}

// GetLogStreams gets the log streams from AWS given the blade configuration
func (b *Blade) GetLogStreams() []types.LogStream {
	input := b.config.DescribeLogStreamsInput()
	streams := make([]types.LogStream, 0)
	streamsOutput, err := b.cwl.DescribeLogStreams(context.TODO(), input)

	if err != nil {
		return streams
	}

	for _, stream := range streamsOutput.LogStreams {
		streams = append(streams, stream)
	}

	return streams
}

// GetEvents gets events from AWS given the blade configuration
func (b *Blade) GetEvents() {
	//formatter := b.output.Formatter()
	input := b.config.FilterLogEventsInput()

	//handlePage := func(page *cloudwatchlogs.FilterLogEventsOutput, lastPage bool) bool {
	//	for _, event := range page.Events {
	//		if b.output.Pretty {
	//			fmt.Println(formatEvent(formatter, &event))
	//		} else {
	//			fmt.Println(*event.Message)
	//		}
	//	}
	//	return !lastPage
	//}
	err, _ := b.cwl.FilterLogEvents(context.TODO(), input)
	if err != nil {
		fmt.Println("Error", err)
		os.Exit(2)
	}
}

// StreamEvents continuously prints log events to the console
func (b *Blade) StreamEvents() {
	//var lastSeenTime *int64
	//var seenEventIDs map[string]bool
	//formatter := b.output.Formatter()
	//input := b.config.FilterLogEventsInput()
	//
	//clearSeenEventIds := func() {
	//	seenEventIDs = make(map[string]bool, 0)
	//}
	//
	//addSeenEventIDs := func(id *string) {
	//	seenEventIDs[*id] = true
	//}
	//
	//updateLastSeenTime := func(ts *int64) {
	//	if lastSeenTime == nil || *ts > *lastSeenTime {
	//		lastSeenTime = ts
	//		clearSeenEventIds()
	//	}
	//}

	//handlePage := func(page *cloudwatchlogs.FilterLogEventsOutput, lastPage bool) bool {
	//	for _, event := range page.Events {
	//		updateLastSeenTime(event.Timestamp)
	//		if _, seen := seenEventIDs[*event.EventId]; !seen {
	//			var message string
	//			if b.output.Raw {
	//				message = *event.Message
	//			} else {
	//				message = formatEvent(formatter, &event)
	//			}
	//			message = strings.TrimRight(message, "\n")
	//			fmt.Println(message)
	//			addSeenEventIDs(event.EventId)
	//		}
	//	}
	//	return !lastPage
	//}

	//for {
	//	err := b.cwl.FilterLogEventsPages(input, handlePage)
	//	if err != nil {
	//		fmt.Println("Error", err)
	//		os.Exit(2)
	//	}
	//	if lastSeenTime != nil {
	//		input.StartTime = lastSeenTime
	//	}
	//	time.Sleep(1 * time.Second)
	//}
}

// formatEvent returns a CloudWatch log event as a formatted string using the provided formatter
func formatEvent(formatter *colorjson.Formatter, event *types.FilteredLogEvent) string {
	red := color.New(color.FgRed).SprintFunc()
	white := color.New(color.FgWhite).SprintFunc()

	str := aws.String(*event.Message)
	bytes := []byte(*str)
	date := MillisecondsTimeValue(event.Timestamp)
	dateStr := date.Format(time.RFC3339)
	streamStr := aws.String(*event.LogStreamName)
	jl := map[string]interface{}{}

	if err := json.Unmarshal(bytes, &jl); err != nil {
		return fmt.Sprintf("[%s] (%s) %s", red(dateStr), white(streamStr), str)
	}

	output, _ := formatter.Marshal(jl)
	return fmt.Sprintf("[%s] (%s) %s", red(dateStr), white(streamStr), output)
}

func MillisecondsTimeValue(v *int64) time.Time {
	if v != nil {
		return time.Unix(0, (*v * 1000000))
	}
	return time.Time{}
}
