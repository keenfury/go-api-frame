package main

import "bufio"

type (
	ProjectFile struct {
		Message              string `json:"message"`
		AppName              string `json:"app_name"`
		FullPath             string `json:"full_path"`    // full path to the project
		SubDir               string `json:"sub_dir"`      // directory, only, you will save the files to
		SubPackage           string `json:"sub_package"`  // if sub directory is multipath i.e. internal/v1, will be used for package name
		ProjectPath          string `json:"project_path"` // fullpath minus gopath/src; used for import statements
		SaveStorage          bool   `json:"save_storage"`
		Storages             string `json:"storages"`             // space delimited string of storage types
		ProjectPathEncoded   string `json:"project_path_encoded"` // encode to use in some of the templating
		SubDirEncoded        string `json:"sub_dir_encoded"`      // encode to use in some of the templating
		DynamicSchema        bool   `json:"dynamic_schema"`
		Schema               string `json:"schema"`
		DynamicSchemaPostfix string `json:"dynamic_schema_postfix"`
		UseORM               bool   `json:"use_orm"`
		Name
	}

	Project struct {
		Reader      *bufio.Reader
		ParseInput  func(*bufio.Reader) string
		UseBlank    bool
		EndPoints   []EndPoint
		ProjectFile ProjectFile
	}

	EndPoint struct {
		Columns                []Column
		ModelIncludeNull       string
		ModelRows              string
		HandlerStrConv         string
		HandlerGetDeleteUrl    string
		HandlerGetDeleteAssign string
		HandlerArgSet          string
		ManagerTime            string
		ManagerGetRow          string
		ManagerPostRows        string
		ManagerPutRows         string
		ManagerPatchRows       string
		ManagerGetTestRow      string
		ManagerPostTestRow     string
		ManagerPutTestRow      string
		ManagerDeleteTestRow   string
		ManagerUtilPath        string
		ManagerImportTest      string
		DataTable              string
		DataTablePrefix        string
		DataTablePostfix       string
		SqlGetColumns          string
		SqlTableKeyKeys        string
		SqlTableKeyValues      string
		SqlTableKeyListOrder   string
		SqlPostColumns         string
		SqlPostColumnsNamed    string
		SqlPostReturning       string
		SqlPostLastId          string
		SqlPatchColumns        string
		SqlPatchWhere          string
		SqlPatchWhereValues    string
		FileKeys               string
		FileGetColumns         string
		FilePostIncr           string
		HaveNullColumns        bool
		SqlLines               Sql
		InitStorage            string // holds the formatted lines for InitStorage for model
		SQLProvider            string // optional if using SQL as a storage, either Psql, MySql or Sqlite; this interfaces with sqlx
		SQLProviderLower       string // optional if using SQL as a storage, either psql, mysql or sqlite; this interfaces with gorm
		SQLProviderConnection  string // holds the connection string for gorm of the other sql types
		MigrationVerify        string
		MigrationConnection    string
		MigrationHeader        string
		GrpcTranslateIn        string
		GrpcTranslateOut       string
		Name
		ProjectFile
	}

	Name struct {
		RawName    string // name given by the user/script
		Lower      string
		Camel      string
		LowerCamel string
		Abbr       string
		AllLower   string
		Upper      string
		EnvVar     string
	}

	Column struct {
		ColumnName   Name
		DBType       string
		GoType       string
		GoTypeNonSql string
		Null         bool
		DefaultValue string
		Length       int64
		PrimaryKey   bool
	}

	Sql struct {
		RawSql []string
	}
)
