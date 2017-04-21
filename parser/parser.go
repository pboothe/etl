// The parser package defines the Parser interface and implementations for the different
// test types, NDT, Paris Traceroute, and SideStream.
package parser

import (
	"log"

	"cloud.google.com/go/bigquery"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/m-lab/etl/intf"
)

//=====================================================================================
//                       Parser implementations
//=====================================================================================
type NullParser struct {
	intf.Parser
}

func (np *NullParser) ParseAndInsert(meta map[string]bigquery.Value, testName string, test []byte) error {
	testCount.With(prometheus.Labels{"table": np.TableName()}).Inc()
	return nil
}

func (np *NullParser) TableName() string {
	return "null-table"
}

type FileNameSaver struct {
	Values map[string]bigquery.Value
}

// TODO(dev) - Figure out if this can use a pointer receiver.
func (fns FileNameSaver) Save() (row map[string]bigquery.Value, insertID string, err error) {
	return fns.Values, "", nil
}

//------------------------------------------------------------------------------------
// TestParser ignores the content, returns a ValueSaver with map[string]Value
// underneath, containing meta data and "testname":"..."
// TODO add tests
type TestParser struct {
	inserter intf.Inserter
	intf.Parser
}

func NewTestParser(ins intf.Inserter) intf.Parser {
	return &TestParser{ins, nil}
}

func (tp *TestParser) ParseAndInsert(meta map[string]bigquery.Value, testName string, test []byte) error {
	testCount.With(prometheus.Labels{"table": tp.TableName()}).Inc()
	log.Printf("Parsing %s", testName)
	values := make(map[string]bigquery.Value, len(meta)+1)
	// TODO is there a better way to do this?
	for k, v := range meta {
		values[k] = v
	}
	values["testname"] = testName
	return tp.inserter.InsertRow(FileNameSaver{values})
}

func (tp *TestParser) TableName() string {
	return "test-table"
}

//=====================================================================================
//                       Prometheus Monitoring
//=====================================================================================

var (
	testCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "etl_parser_test_count",
		Help: "Number of tests processed.",
	}, []string{"table"})

	failureCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "etl_parser_failure_count",
		Help: "Number of test processing failures.",
	}, []string{"table", "failure_type"})
)

func init() {
	prometheus.MustRegister(testCount)
	prometheus.MustRegister(failureCount)
}
