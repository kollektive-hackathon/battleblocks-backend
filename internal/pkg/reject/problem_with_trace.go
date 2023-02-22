package reject

type ProblemWithTrace struct {
	Problem Problem
	Cause   error
}
