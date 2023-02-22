package configuration

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xeipuuv/gojsonschema"
)

func TestConfigurationSchemaValidity(t *testing.T) {
	schemaBytes, err := ioutil.ReadFile("configuration.schema.json")
	require.NoError(t, err)

	jl := gojsonschema.NewBytesLoader(schemaBytes)
	sl := gojsonschema.NewSchemaLoader()
	_, err = sl.Compile(jl)
	require.NoError(t, err)
}
