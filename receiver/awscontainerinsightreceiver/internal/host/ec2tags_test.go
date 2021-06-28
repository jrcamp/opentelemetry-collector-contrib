// Copyright  OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package host

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/awstesting/mock"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type mockEC2TagsClient struct {
	count int
}

var tokenString = "tokenString"
var clusterKey = clusterNameTagKeyPrefix + "cluster-name"
var clusterValue = "owned"
var asgKey = autoScalingGroupNameTag
var asgValue = "asg"

func (m *mockEC2TagsClient) DescribeTagsWithContext(ctx context.Context, input *ec2.DescribeTagsInput,
	opts ...request.Option) (*ec2.DescribeTagsOutput, error) {
	m.count++
	if m.count == 1 {
		return &ec2.DescribeTagsOutput{}, errors.New("error")
	}

	if m.count == 2 {
		return &ec2.DescribeTagsOutput{
			NextToken: &tokenString,
			Tags: []*ec2.TagDescription{
				{
					Key:   &asgKey,
					Value: &asgValue,
				},
			},
		}, nil
	}

	return &ec2.DescribeTagsOutput{
		Tags: []*ec2.TagDescription{
			{
				Key:   &clusterKey,
				Value: &clusterValue,
			},
		},
	}, nil
}

func TestEC2Tags(t *testing.T) {
	ctx := context.Background()
	sess := mock.Session
	clientOption := func(e *ec2Tags) {
		e.client = &mockEC2TagsClient{}
	}
	maxJitterOption := func(e *ec2Tags) {
		e.maxJitterTime = 0
	}
	isSucessOption := func(e *ec2Tags) {
		e.isSucess = make(chan bool)
	}
	et := newEC2Tags(ctx, sess, "instanceId", "us-west-2", time.Millisecond, zap.NewNop(), clientOption,
		maxJitterOption, isSucessOption)

	// wait for ec2 tags are fetched
	e := et.(*ec2Tags)
	<-e.isSucess
	assert.Equal(t, "cluster-name", et.getClusterName())
	assert.Equal(t, "asg", et.getAutoScalingGroupName())
}