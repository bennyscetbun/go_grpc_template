package main

import (
	"database/sql"
	"flag"
	"log"
	"path/filepath"

	"github.com/bennyscetbun/xxx_your_app_xxx/backend/cmd/database_gen/generatehelpers"
	"github.com/bennyscetbun/xxx_your_app_xxx/backend/database"
	"gorm.io/driver/postgres"
	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func generateCode(db *sql.DB, generatedFilesPath string) error {
	g := gen.NewGenerator(gen.Config{
		OutPath:          filepath.Join(generatedFilesPath, "dbqueries"),
		ModelPkgPath:     filepath.Join(generatedFilesPath, "dbmodels"),
		Mode:             gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface, // generate mode
		FieldNullable:    true,
		FieldWithTypeTag: true,
	})

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return err
	}

	g.UseDB(gormDB)

	tableList, err := gormDB.Migrator().GetTables()
	if err != nil {
		return err
	}
	tableModels := make([]interface{}, 0, len(tableList))
	for _, tableName := range tableList {
		if tableName == "schema_migrations" {
			continue
		}
		types, err := gormDB.Migrator().ColumnTypes(tableName)
		if err != nil {
			return err
		}
		hasUpdatedAt := false
		hasCreatedAt := false
		for _, col := range types {
			switch col.Name() {
			case "updated_at":
				hasUpdatedAt = true
			case "created_at":
				hasCreatedAt = true
			}
		}
		if hasCreatedAt && hasUpdatedAt {
			tableModels = append(tableModels, g.GenerateModel(tableName, gen.WithMethod(generatehelpers.CreatedAtUpdatedAtAble{})))
		} else if hasCreatedAt {
			tableModels = append(tableModels, g.GenerateModel(tableName, gen.WithMethod(generatehelpers.CreatedAtAble{})))
		} else if hasUpdatedAt {
			tableModels = append(tableModels, g.GenerateModel(tableName, gen.WithMethod(generatehelpers.UpdatedAtAble{})))
		} else {
			tableModels = append(tableModels, g.GenerateModel(tableName))
		}
	}

	g.ApplyBasic(tableModels...)
	g.Execute()
	return nil
}

func main() {
	generatedFilesPath := flag.String("out", "./generated/database", "generated files path")
	flag.Parse()

	db, err := database.OpenPSQL()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = database.MigratePSQL(db)
	if err != nil {
		log.Fatal(err)
	}

	err = generateCode(db, *generatedFilesPath)
	if err != nil {
		log.Fatal(err)
	}
}
