// THIS FILE IS AUTOMATICALLY GENERATED. DO NOT EDIT.

package elbiface_test

import (
	"testing"

	"github.com/convox/rack/Godeps/_workspace/src/github.com/aws/aws-sdk-go/service/elb"
	"github.com/convox/rack/Godeps/_workspace/src/github.com/aws/aws-sdk-go/service/elb/elbiface"
	"github.com/convox/rack/Godeps/_workspace/src/github.com/stretchr/testify/assert"
)

func TestInterface(t *testing.T) {
	assert.Implements(t, (*elbiface.ELBAPI)(nil), elb.New(nil))
}