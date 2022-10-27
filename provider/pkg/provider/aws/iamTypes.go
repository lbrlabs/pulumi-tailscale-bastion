package aws

type AssumeRolePolicy struct {
	Version   string      `json:"Version"`
	Statement []Statement `json:"Statement"`
}
type Principal struct {
	Service []string `json:"Service"`
}
type Statement struct {
	Effect    string    `json:"Effect"`
	Principal Principal `json:"Principal"`
	Action    string    `json:"Action"`
}
