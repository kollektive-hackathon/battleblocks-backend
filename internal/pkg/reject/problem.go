package reject

type Problem struct {
	Title  string            `json:"title,omitempty"`
	Status int               `json:"status,omitempty"`
	Detail string            `json:"detail,omitempty"`
	Code   string            `json:"message,omitempty"`
	Type   string            `json:"type,omitempty"`
	Path   string            `json:"path,omitempty"`
	Params map[string]string `json:"params,omitempty"`
	Errors []ProblemDetail   `json:"errors,omitempty"`
}

type ProblemDetail struct {
	Property string `json:"property,omitempty"`
	Info     string `json:"info,omitempty"`
	Code     string `json:"code,omitempty"`
}

func NewProblem() *Problem {
	return &Problem{}
}

func (p *Problem) WithTitle(title string) *Problem {
	p.Title = title
	return p
}

func (p *Problem) WithStatus(status int) *Problem {
	p.Status = status
	return p
}

func (p *Problem) WithDetail(detail string) *Problem {
	p.Detail = detail
	return p
}

func (p *Problem) WithCode(code string) *Problem {
	p.Code = code
	return p
}

func (p *Problem) WithType(typeValue string) *Problem {
	p.Type = typeValue
	return p
}

func (p *Problem) WithPath(path string) *Problem {
	p.Path = path
	return p
}

func (p *Problem) WithParam(key string, value string) *Problem {
	p.Params[key] = value
	return p
}

func (p *Problem) WithErrors(errors []ProblemDetail) *Problem {
	p.Errors = errors
	return p
}

func (p *Problem) Build() Problem {
	return *p
}
