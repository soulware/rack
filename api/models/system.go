package models

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/convox/rack/Godeps/_workspace/src/github.com/aws/aws-sdk-go/aws"
	"github.com/convox/rack/Godeps/_workspace/src/github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/convox/rack/Godeps/_workspace/src/github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/convox/rack/client"
)

type System client.System

func GetSystem() (*System, error) {
	rack := os.Getenv("RACK")

	res, err := CloudFormation().DescribeStacks(&cloudformation.DescribeStacksInput{StackName: aws.String(rack)})

	if err != nil {
		return nil, err
	}

	if len(res.Stacks) != 1 {
		return nil, fmt.Errorf("could not load stack for app: %s", rack)
	}

	stack := res.Stacks[0]
	params := stackParameters(stack)

	count, err := strconv.Atoi(params["InstanceCount"])

	if err != nil {
		return nil, err
	}

	r := &System{
		Count:   count,
		Name:    rack,
		Status:  humanStatus(*stack.StackStatus),
		Type:    params["InstanceType"],
		Version: os.Getenv("RELEASE"),
	}

	return r, nil
}

func (r *System) Save() error {
	rack := os.Getenv("RACK")

	app, err := GetApp(rack)

	if err != nil {
		return err
	}

	mac, err := maxAppConcurrency()

	// dont scale the rack below the max concurrency plus one
	// see formation.go for more details
	if err == nil && r.Count < (mac+1) {
		return fmt.Errorf("max process concurrency is %d, can't scale rack below %d instances", mac, mac+1)
	}

	params := map[string]string{
		"InstanceCount": strconv.Itoa(r.Count),
		"InstanceType":  r.Type,
		"Version":       r.Version,
	}

	template := fmt.Sprintf("https://convox.s3.amazonaws.com/release/%s/formation.json", r.Version)

	err = app.UpdateParamsAndTemplate(params, template)

	if err != nil {
		return err
	}

	// On version update, create an app release record
	if app.Outputs["Release"] == r.Version {
		return nil
	}

	release, err := app.LatestRelease()

	if err != nil {
		return err
	}

	if release == nil {
		r := NewRelease(app.Name)
		release = &r
	}

	release.Id = r.Version
	release.Created = time.Now()

	req := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"id":      &dynamodb.AttributeValue{S: aws.String(release.Id)},
			"app":     &dynamodb.AttributeValue{S: aws.String(release.App)},
			"created": &dynamodb.AttributeValue{S: aws.String(release.Created.Format(SortableTime))},
		},
		TableName: aws.String(releasesTable(release.App)),
	}

	_, err = DynamoDB().PutItem(req)

	return err
}

func maxAppConcurrency() (int, error) {
	apps, err := ListApps()

	if err != nil {
		return 0, err
	}

	max := 0

	for _, app := range apps {
		rel, err := app.LatestRelease()

		if err != nil {
			return 0, err
		}

		if rel == nil {
			continue
		}

		m, err := LoadManifest(rel.Manifest)

		if err != nil {
			return 0, err
		}

		f, err := ListFormation(app.Name)

		if err != nil {
			return 0, err
		}

		for _, me := range m {
			if len(me.ExternalPorts()) > 0 {
				entry := f.Entry(me.Name)

				if entry != nil && entry.Count > max {
					max = entry.Count
				}
			}
		}
	}

	return max, nil
}
