package config_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	xcfg "x.io/xrpc/pkg/config"
)

var cfg = xcfg.NewConfig()

func TestXCfg(t *testing.T) {
	xcfg.StringVar("system", "xcfg", "system name")
	xcfg.StringVar("config", "./test.json", "json config path")
	xcfg.LoadEnv([]string{"GOPATH", "GOROOT"})
	xcfg.Parse()
	cfg := xcfg.String("config")
	xcfg.LoadCfg(cfg)

	fmt.Println(xcfg.String("system"))       // xcfg
	fmt.Println(xcfg.String("GOROOT"))       // /opt/soft/go
	fmt.Println(xcfg.IntMap("maps.int"))     // map[i:1]
	fmt.Println(xcfg.FloatMap("maps.float")) // map[f:0.01]
	fmt.Println(xcfg.StrMap("maps.str"))     //map[a:a]
	fmt.Println(xcfg.IntArray("int_arr"))    // [0 1 2 3 4 5 6 7 8 9 10 11]
	fmt.Println(xcfg.Int("int_arr.1"))       // 1

	pattern := "int_arr.[^1234567]"
	fmt.Println(xcfg.Match(pattern).IntMap()) // map[int_arr.0:0 int_arr.9:9 int_arr.8:8]
	m := xcfg.Map("maps.int")
	fmt.Println(m["i"].Int())
}

func TestConfigDefault(t *testing.T) {
	project := "xcfg"
	xcfg.SetDefaultKey("project", &project)
	assert.Equal(t, project, cfg.String("project"))
}

func TestConfigFlag(t *testing.T) {
	var name = "haha"
	var age = 21
	var phone = int64(17877652365)
	var id uint64 = 192839
	var money = 0.01

	xcfg.StringVar("name", name, "my name")
	xcfg.IntVar("age", age, "Age")
	xcfg.Int64Var("phone", phone, "Phone number")
	xcfg.Uint64Var("id", id, "ID number")
	xcfg.Float64Var("money", money, "Left money")
	xcfg.Parse()

	assert.Equal(t, name, cfg.String("name"))
	assert.Equal(t, age, cfg.Int("age"))
	assert.Equal(t, phone, cfg.Int64("phone"))
	assert.Equal(t, id, cfg.Uint64("id"))
	assert.Equal(t, money, cfg.Float("money"))
}

func TestConfigJson(t *testing.T) {
	// default config path is "./config.json"
	cfg.LoadJsonCfg("./test.json")
	assert.Equal(t, "./test.json", cfg.String("config"))
	assert.Equal(t, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}, cfg.IntArray("int_arr"))
	assert.Equal(t, []float64{0.1, 0.2, 0.3}, cfg.FloatArray("float_arr"))
	assert.Equal(t, "INetwork", cfg.String("network.name"))
	assert.Equal(t, "Boom", cfg.String("A.B.C"))
}

func BenchmarkConfigJson(b *testing.B) {
	// default config path is "./config.json"
	cfg.LoadJsonCfg("./test.json")
	for i := 0; i < b.N; i++ {
		assert.Equal(b, []float64{0.1, 0.2, 0.3}, cfg.FloatArray("float_arr"))
	}
}

func TestConfigEnv(t *testing.T) {
	keys := []string{"GOROOT", "GOPATH", "PATH"}
	xcfg.LoadEnv(keys)
	assert.Equal(t, "/usr/local/go", cfg.String("GOROOT"))
}

func TestConfig_Match(t *testing.T) {
	cfg.LoadJsonCfg("./test.json")
	pattern := `match.sub_[a-z]*.[a-z]a[a-z]`
	assert.Equal(t, map[string]int{
		"match.sub_match.cab": 3,
		"match.sub_match.bac": 2,
	}, cfg.Match(pattern).IntMap())
	pattern = `network.listeners.[0-9].protocol`
	except := map[string]string{
		"network.listeners.0.protocol": "udp",
		"network.listeners.1.protocol": "tcp",
		"network.listeners.2.protocol": "kcp",
	}
	assert.Equal(t, except, cfg.Match(pattern).StrMap())

	pattern = `int_arr.[0-9]{2}`
	intExcept := map[string]int{
		"int_arr.10": 10,
		"int_arr.11": 11,
	}
	assert.Equal(t, intExcept, cfg.Match(pattern).IntMap())
}

func TestConfig_Dump(t *testing.T) {
	if false {
		cfg.LoadJsonCfg("./test.json")
		peerName := "NewPeer"
		cfg.Set("peer_name", &peerName)
		assert.Equal(t, nil, cfg.Dump())
		// Check new config
		cfg.LoadJsonCfg()
		assert.Equal(t, peerName, cfg.String("peer_name"))
	}
}
