package workflows

import (
	"easyRide/activities"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"
	"testing"
	"time"
)

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	env *testsuite.TestWorkflowEnvironment
}

func (s *UnitTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

func (s *UnitTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func (s *UnitTestSuite) Test_MainWorkflow_Success() {
	s.env.OnActivity(activities.InTrip, mock.Anything, mock.Anything).Return(nil)
	s.env.OnActivity(activities.Arrive, mock.Anything, mock.Anything).Return(nil)
	s.env.OnActivity(activities.PassengerEndTrip, mock.Anything, mock.Anything).Return(nil)
	s.env.OnActivity(activities.Rate, mock.Anything, mock.Anything).Return(nil)

	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("signal_match", true)
	}, time.Millisecond*1)

	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("signal_payment", true)
	}, time.Millisecond*3)

	s.env.ExecuteWorkflow(MainWorkFlow, 1)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}
