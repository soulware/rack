package main

import (
	"net/http/httptest"
	"testing"

	"github.com/convox/rack/Godeps/_workspace/src/github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/convox/rack/Godeps/_workspace/src/github.com/convox/release/version"
	"github.com/convox/rack/Godeps/_workspace/src/github.com/stretchr/testify/assert"
	"github.com/convox/rack/api/awsutil"
	"github.com/convox/rack/test"
)

func TestConvoxInstallSTDINCredentials(t *testing.T) {
	stackId := "arn:aws:cloudformation:us-east-1:123456789:stack/MyStack/aaf549a0-a413-11df-adb3-5081b3858e83"
	cycles := []awsutil.Cycle{
		awsutil.Cycle{
			awsutil.Request{"/", "", "/./"},
			awsutil.Response{200, `<CreateStackResult><StackId>` + stackId + `</StackId></CreateStackResult>`},
		},
		awsutil.Cycle{
			awsutil.Request{"/", "", ""},
			awsutil.Response{200, ""},
		},
	}

	handler := awsutil.NewHandler(cycles)
	s := httptest.NewServer(handler)
	defaults.DefaultConfig.Endpoint = &s.URL

	defer s.Close()

	latest, _ := version.Latest()

	test.Runs(t,
		test.ExecRun{
			Command: "convox install",
			Exit:    0,
			Env:     map[string]string{"AWS_ENDPOINT_URL": s.URL, "AWS_REGION": "test"},
			Stdin:   `{"Credentials":{"AccessKeyId":"FOO","SecretAccessKey":"BAR","Expiration":"2015-09-17T14:09:41Z"}}`,
			Stdout:  Banner + "\nInstalling Convox (" + latest + ")...\n" + stackId + "\n",
		},
	)
}

func TestConvoxInstallFileCredentials(t *testing.T) {

}

func TestConvoxInstallSubnetCalculation(t *testing.T) {
	// test some invalid CIDRs
	tooSmallCIDR := "10.0.0.0/28"
	subnets, err := calculateSubnets(tooSmallCIDR)
	assert.Equal(t, "VPC CIDR must be between /16 and /26", err.Error())
	assert.Equal(t, []string(nil), subnets)

	tooLargeCIDR := "10.0.0.0/8"
	subnets, err = calculateSubnets(tooLargeCIDR)
	assert.Equal(t, "VPC CIDR must be between /16 and /26", err.Error())
	assert.Equal(t, []string(nil), subnets)

	invalidCIDR := "10.0.0.0"
	subnets, err = calculateSubnets(invalidCIDR)
	assert.Equal(t, "invalid CIDR address: 10.0.0.0", err.Error())
	assert.Equal(t, []string(nil), subnets)

	// test some valid CIDRs

	// largest valid CIDR
	validCIDR16 := "10.0.0.0/16"
	subnets, err = calculateSubnets(validCIDR16)
	assert.Equal(t, nil, err)
	assert.Equal(t, []string{"10.0.0.0/18", "10.0.64.0/18", "10.0.128.0/18"}, subnets)

	// smallest valid CIDR
	validCIDR27 := "10.0.0.0/27"
	subnets, err = calculateSubnets(validCIDR27)
	assert.Equal(t, nil, err)
	assert.Equal(t, []string{"10.0.0.0/29", "10.0.0.8/29", "10.0.0.16/29"}, subnets)

	validCIDR24 := "10.0.0.0/24"
	subnets, err = calculateSubnets(validCIDR24)
	assert.Equal(t, nil, err)
	assert.Equal(t, []string{"10.0.0.0/26", "10.0.0.64/26", "10.0.0.128/26"}, subnets)

	validCIDR17 := "10.0.0.0/17"
	subnets, err = calculateSubnets(validCIDR17)
	assert.Equal(t, nil, err)
	assert.Equal(t, []string{"10.0.0.0/19", "10.0.32.0/19", "10.0.64.0/19"}, subnets)
}
