package db

type Implementation interface {
	UserQuery() UserQuery
	RoleQuery() RoleQuery
}

type implementation struct {
	userQuery UserQuery
	roleQuery RoleQuery
}

func NewImplementation(userQuery UserQuery, roleQuery RoleQuery) Implementation {
	return &implementation{
		userQuery: userQuery,
		roleQuery: roleQuery,
	}
}

func (i *implementation) UserQuery() UserQuery {
	return i.userQuery
}

func (i *implementation) RoleQuery() RoleQuery {
	return i.roleQuery
}
