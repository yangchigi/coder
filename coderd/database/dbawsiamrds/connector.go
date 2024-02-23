package dbawsiamrds

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds/rdsutils"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"golang.org/x/xerrors"
)

var DriverNotSupportedErr = xerrors.New("driver open method not supported")

var _ driver.Connector = &AwsIamConnector{}
var _ driver.Driver = &AwsIamConnector{}

type AwsIamConnector struct {
	sess  *session.Session
	dbURL string
}

// NewDB will create a new *sqlx.DB using the environment aws session and ping postgres.
func NewDB(ctx context.Context, dbURL string) (*sqlx.DB, error) {
	c, err := NewConnector(dbURL)
	if err != nil {
		return nil, xerrors.Errorf("creating connector: %w", err)
	}

	sqlDB := sql.OpenDB(c)
	sqlxDB := sqlx.NewDb(sqlDB, "postgres")

	err = sqlxDB.PingContext(ctx)
	if err != nil {
		return nil, xerrors.Errorf("ping postgres: %w", err)
	}

	return sqlxDB, nil
}

// NewConnector will create a new *AwsIamConnector using the environment aws session.
func NewConnector(dbURL string) (*AwsIamConnector, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, xerrors.Errorf("creating aws session: %w", err)
	}

	c := &AwsIamConnector{
		sess:  sess,
		dbURL: dbURL,
	}

	return c, nil
}

// Connect fulfills the driver.Connector interface using aws iam rds credentials from the environment
func (c *AwsIamConnector) Connect(ctx context.Context) (driver.Conn, error) {
	// set password with signed aws authentication token for the rds instance
	nURL, err := GetAuthenticatedURL(c.sess, c.dbURL)
	if err != nil {
		return nil, xerrors.Errorf("assigning authentication token to url: %w", err)
	}

	fmt.Println(nURL)
	// make connection
	connector, err := pq.NewConnector(nURL)
	if err != nil {
		return nil, xerrors.Errorf("building new pq connector: %w", err)
	}

	conn, err := connector.Connect(ctx)
	if err != nil {
		return nil, xerrors.Errorf("making connection: %w", err)
	}

	return conn, nil
}

// Driver fulfills the driver.Connector interface.
func (c *AwsIamConnector) Driver() driver.Driver {
	return c
}

// Open fulfills the driver.Driver interface with an error.
// This interface should not be opened via the driver open method.
func (_ *AwsIamConnector) Open(_ string) (driver.Conn, error) {
	return nil, DriverNotSupportedErr
}

func GetAuthenticatedURL(sess *session.Session, dbURL string) (string, error) {
	nURL, err := url.Parse(dbURL)
	if err != nil {
		return "", xerrors.Errorf("parsing dbURL: %w", err)
	}

	// generate a new rds session auth tokenized URL
	rdsEndpoint := fmt.Sprintf("%s:%s", nURL.Hostname(), nURL.Port())
	token, err := rdsutils.BuildAuthToken(rdsEndpoint, *sess.Config.Region, nURL.User.Username(), sess.Config.Credentials)
	if err != nil {
		return "", xerrors.Errorf("building rds auth token: %w", err)
	}
	// set token as user password
	nURL.User = url.UserPassword(nURL.User.Username(), token)

	return nURL.String(), nil
}
