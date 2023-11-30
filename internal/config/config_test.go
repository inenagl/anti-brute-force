package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestConfigFromFile(t *testing.T) {
	c, err := New("testdata/config.yaml", "")
	require.NoError(t, err)
	require.Equal(t, 1, c.Main.MaxLogins)
	require.Equal(t, 2, c.Main.MaxPasswords)
	require.Equal(t, 3, c.Main.MaxIPs)
	require.Equal(t, 4, c.Main.CacheSize)
	require.Equal(t, time.Second, c.Main.CacheTTL)
	require.Equal(t, time.Second*2, c.Main.BucketTTL)

	require.Equal(t, "dev", c.Logger.Preset)
	require.Equal(t, "debug", c.Logger.Level)
	require.Equal(t, "console", c.Logger.Encoding)
	require.Equal(t, []string{"/logfile.log"}, c.Logger.OutputPaths)
	require.Equal(t, []string{"/errorfile.log", "stdout"}, c.Logger.ErrorOutputPaths)

	require.Equal(t, "testdbhost", c.DB.Host)
	require.Equal(t, 2222, c.DB.Port)
	require.Equal(t, "testuser", c.DB.User)
	require.Equal(t, "testpassword", c.DB.Password)
	require.Equal(t, "test", c.DB.SSLMode)
	require.Equal(t, time.Second, c.DB.Timeout)
	require.Equal(t, "testdb", c.DB.DBName)

	require.Equal(t, "testapihost", c.APIServer.Host)
	require.Equal(t, 1111, c.APIServer.Port)
}

func TestConfigFromFileAndEnv(t *testing.T) {
	setEnv(t, "TST", true)
	c, err := New("testdata/config.yaml", "TST")
	require.NoError(t, err)
	require.Equal(t, 1000, c.Main.MaxLogins)
	require.Equal(t, 2000, c.Main.MaxPasswords)
	require.Equal(t, 3000, c.Main.MaxIPs)
	require.Equal(t, 4, c.Main.CacheSize)
	require.Equal(t, time.Second, c.Main.CacheTTL)
	require.Equal(t, time.Second*2, c.Main.BucketTTL)

	require.Equal(t, "dev", c.Logger.Preset)
	require.Equal(t, "debug", c.Logger.Level)
	require.Equal(t, "console", c.Logger.Encoding)
	require.Equal(t, []string{"/logfile.log"}, c.Logger.OutputPaths)
	require.Equal(t, []string{"/errorfile.log", "stdout"}, c.Logger.ErrorOutputPaths)

	require.Equal(t, "envhost", c.DB.Host)
	require.Equal(t, 4444, c.DB.Port)
	require.Equal(t, "envuser", c.DB.User)
	require.Equal(t, "envpassword", c.DB.Password)
	require.Equal(t, "envssl", c.DB.SSLMode)
	require.Equal(t, time.Hour, c.DB.Timeout)
	require.Equal(t, "envdbname", c.DB.DBName)

	require.Equal(t, "testapihost", c.APIServer.Host)
	require.Equal(t, 1111, c.APIServer.Port)
}

func TestDefaults(t *testing.T) {
	setEnv(t, "DEF", false)

	c, err := New("testdata/empty_config.yaml", "DEF")
	require.NoError(t, err)
	require.Equal(t, defaultMaxLogins, c.Main.MaxLogins)
	require.Equal(t, defaultMaxPasswords, c.Main.MaxPasswords)
	require.Equal(t, defaultMaxIPs, c.Main.MaxIPs)
	require.Equal(t, defaultCacheSize, c.Main.CacheSize)
	require.Equal(t, defaultCacheTTL, c.Main.CacheTTL)
	require.Equal(t, defaultBucketTTL, c.Main.BucketTTL)

	require.Equal(t, defaultLogPreset, c.Logger.Preset)
	require.Equal(t, defaultLogLevel, c.Logger.Level)
	require.Equal(t, defaultLogEncoding, c.Logger.Encoding)
	require.Equal(t, defaultLogOutputPaths, c.Logger.OutputPaths)
	require.Equal(t, defaultLogErrorOutputPaths, c.Logger.ErrorOutputPaths)

	require.Equal(t, defaultAPIServerHost, c.APIServer.Host)
	require.Equal(t, defaultAPIServerPort, c.APIServer.Port)
}

func setEnv(t *testing.T, prefix string, withMain bool) {
	t.Helper()

	err := os.Setenv(prefix+"_DBHOST", "envhost")
	require.NoError(t, err)
	err = os.Setenv(prefix+"_DBPORT", "4444")
	require.NoError(t, err)
	err = os.Setenv(prefix+"_DBUSER", "envuser")
	require.NoError(t, err)
	err = os.Setenv(prefix+"_DBPASSWORD", "envpassword")
	require.NoError(t, err)
	err = os.Setenv(prefix+"_DBSSLMODE", "envssl")
	require.NoError(t, err)
	err = os.Setenv(prefix+"_DBTIMEOUT", "1h")
	require.NoError(t, err)
	err = os.Setenv(prefix+"_DBNAME", "envdbname")
	require.NoError(t, err)

	if withMain {
		err = os.Setenv(prefix+"_MAIN_MAXLOGINS", "1000")
		require.NoError(t, err)
		err = os.Setenv(prefix+"_MAIN_MAXPASSWORDS", "2000")
		require.NoError(t, err)
		err = os.Setenv(prefix+"_MAIN_MAXIPS", "3000")
		require.NoError(t, err)
	}
}
