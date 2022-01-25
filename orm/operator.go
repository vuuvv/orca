package orm

const (
	OP_Equal              string = "="
	OP_NotEqual           string = "!="
	OP_NotEqualAnother    string = "<>"
	OP_GreaterThan        string = ">"
	OP_GreaterThanOrEqual string = ">="
	OP_LessThan           string = "<"
	OP_LessThanOrEqual    string = "<="
	OP_In                 string = "IN"
	OP_NotIn              string = "NOT IN"
	OP_Like               string = "LIKE"
	OP_NotLike            string = "NOT LIKE"
	OP_Between            string = "BETWEEN"
	OP_NotBetween         string = "NOT BETWEEN"
	OP_Contain            string = "CONTAIN"
	OP_StartsWith         string = "STARTS_WITH"
	OP_EndsWith           string = "ENDS_WITH"
)
