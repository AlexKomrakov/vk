package main

import (
    "time"
    "fmt"
    "testing"
    vegeta "github.com/tsenart/vegeta/lib"
)


func TestMain(t *testing.T) {
    rate := uint64(3000) // per second
    duration := 30 * time.Second
    targeter := vegeta.NewStaticTargeter(vegeta.Target{
        Method: "GET",
        URL:    "http://api.vk.com/method/users.get?user_id=1&v=5.37",
    })
    attacker := vegeta.NewAttacker(vegeta.Timeout(5 * time.Second))

    var results vegeta.Results
    for res := range attacker.Attack(targeter, rate, duration) {
        results = append(results, res)
    }

    metrics := vegeta.NewMetrics(results)
    fmt.Println(metrics)
    fmt.Printf("99th percentile: %s\n", metrics.Latencies.P99)
}
