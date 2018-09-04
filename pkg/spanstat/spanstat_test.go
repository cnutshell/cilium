// Copyright 2018 Authors of Cilium
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package spanstat

import (
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	TestingT(t)
}

type SpanStatTestSuite struct{}

var _ = Suite(&SpanStatTestSuite{})

func (s *SpanStatTestSuite) TestSpanStat(c *C) {
	span1 := SpanStat{}

	// no spans measured yet
	c.Assert(span1.Total(), Equals, time.Duration(0))

	// End() without Start()
	span1.End()
	c.Assert(span1.Total(), Equals, time.Duration(0))

	// Start() but no end yet
	span1.Start()
	c.Assert(span1.Total(), Equals, time.Duration(0))

	// First span measured with End()
	span1.End()
	firstSpanTotal := span1.Total()
	c.Assert(firstSpanTotal, Not(Equals), time.Duration(0))

	// End() without a prior Start(), no change
	span1.End()
	c.Assert(span1.Total(), Equals, firstSpanTotal)

	span1.Start()
	span1.End()
	c.Assert(span1.Total(), Not(Equals), firstSpanTotal)
	c.Assert(span1.Total(), Not(Equals), time.Duration(0))

}
