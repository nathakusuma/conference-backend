package enum

type UserRole string

const (
	RoleSuperAdmin       UserRole = "super_admin"
	RoleAdmin            UserRole = "admin"
	RoleUser             UserRole = "user"
	RoleEventCoordinator UserRole = "event_coordinator"
)

func (r UserRole) String() string {
	return string(r)
}
