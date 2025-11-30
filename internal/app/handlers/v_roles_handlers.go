package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"adminbe/internal/app/models"

	"github.com/gin-gonic/gin"
)

// listVRolesHandler GET /api/v_roles
func listVRolesHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query("SELECT role_id, role_name, child_id, child_name, level FROM v_roles")
		if err != nil {
			log.Printf("Error querying v_roles: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve role hierarchies"})
			return
		}
		defer rows.Close()

		var vRoles []models.VRole
		for rows.Next() {
			var vr models.VRole
			if err := rows.Scan(&vr.RoleID, &vr.RoleName, &vr.ChildID, &vr.ChildName, &vr.Level); err != nil {
				log.Printf("Error scanning v_role row: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve role hierarchies"})
				return
			}
			vRoles = append(vRoles, vr)
		}
		c.JSON(http.StatusOK, gin.H{"data": vRoles})
	}
}
