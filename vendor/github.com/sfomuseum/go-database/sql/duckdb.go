package sql

import (
	"context"
	"fmt"
	db_sql "database/sql"
)

// LoadDuckDBExtensions will issue 'INSTALL' and 'LOAD' statements for 'extensions' using 'db'.
func LoadDuckDBExtensions(ctx context.Context, db *db_sql.DB, extensions ...string) error {

	for _, ext := range extensions {
		
		commands := []string{
			fmt.Sprintf("INSTALL %s", ext),
			fmt.Sprintf("LOAD %s", ext),
		}
		
		for _, cmd := range commands {
			
			_, err := db.ExecContext(ctx, cmd)
			
			if err != nil {
				return fmt.Errorf("Failed to issue command for extension '%s', %w", cmd, err)
			}
		}
	}
	
	return nil
}
