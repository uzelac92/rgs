package observability

import "github.com/prometheus/client_golang/prometheus"

var (
	HttpRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rgs_http_requests_total",
			Help: "Total number of HTTP requests labeled by method, path, and status",
		},
		[]string{"method", "path", "status"},
	)

	BetsPlaced = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "rgs_bets_placed_total",
		Help: "Total number of bets placed",
	})

	WalletDebitCalls = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "rgs_wallet_debit_calls_total",
		Help: "How many debit() calls to wallet were made",
	})

	WalletDebitFailures = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "rgs_wallet_debit_failures_total",
		Help: "How many debit() calls failed",
	})

	BetSettlementDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "rgs_bet_settlement_seconds",
		Help:    "Time spent processing bet settlement pipeline",
		Buckets: prometheus.DefBuckets,
	})
)

func InitMetrics() {
	prometheus.MustRegister(HttpRequests)
	prometheus.MustRegister(BetsPlaced)
	prometheus.MustRegister(WalletDebitCalls)
	prometheus.MustRegister(WalletDebitFailures)
	prometheus.MustRegister(BetSettlementDuration)
}
