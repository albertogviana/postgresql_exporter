package gauges

import (
	"time"

	"github.com/apex/log"
	"github.com/prometheus/client_golang/prometheus"
)

type Relation struct {
	Name string `db:"relname"`
}

var relationsQuery = `
SELECT relname
FROM pg_stat_user_tables
ORDER BY n_tup_ins + n_tup_upd desc
LIMIT 20
`

func (g *Gauges) DeadTuples() *prometheus.GaugeVec {
	var gauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:        "postgresql_dead_tuples_pct",
		Help:        "dead tuples percentage on the top 20 biggest tables",
		ConstLabels: g.labels,
	}, []string{"table"})

	if !g.isSuperuser {
		log.Warn("postgresql_dead_tuples_pct disabled because pgstattuple requires a superuser")
		return gauge
	}
	if !g.hasExtension("pgstattuple") {
		log.Warn("postgresql_dead_tuples_pct disabled because pgstattuple extension is not installed")
		return gauge
	}

	go func() {
		for {
			var tables []Relation
			g.query(relationsQuery, &tables, emptyParams)
			for _, table := range tables {
				var pct []float64
				if err := g.queryWithTimeout(
					"SELECT dead_tuple_percent FROM pgstattuple($1)",
					&pct,
					[]interface{}{table.Name},
					1*time.Minute,
				); err == nil {
					gauge.With(prometheus.Labels{"table": table.Name}).Set(pct[0])
				}
			}
			time.Sleep(12 * time.Hour)
		}
	}()

	return gauge
}
