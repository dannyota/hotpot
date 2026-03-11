package volume

import (
	"context"
	"database/sql"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	"danny.vn/hotpot/pkg/admin"
	lh "danny.vn/hotpot/pkg/admin/listhandler"
	entvolume "danny.vn/hotpot/pkg/storage/ent/greennode/volume"
	pvol "danny.vn/hotpot/pkg/storage/ent/greennode/volume/bronzegreennodevolumeblockvolume"
	"danny.vn/hotpot/pkg/storage/ent/greennode/volume/predicate"
)

// Register registers all GreenNode Volume admin routes.
func Register(driver dialect.Driver, db *sql.DB) {
	entClient := entvolume.NewClient(
		entvolume.Driver(driver),
		entvolume.AlternateSchema(entvolume.DefaultSchemaConfig()),
	)

	filterOpts := &lh.FilterOptionsConfig{
		DB:      db,
		Schema:  "bronze",
		Table:   "greennode_volume_block_volumes",
		Columns: []string{"status", "region"},
	}

	admin.RegisterRoute(admin.RouteRegistration{
		Method: "GET",
		Path:   "/api/v1/bronze/greennode/volume/block-volumes",
		Nav:    &admin.NavMeta{Label: "Block Volumes", Group: []string{"Bronze", "GreenNode", "Volume"}},
		Handler: lh.Handler(lh.Config{
			EntityName: "block-volumes",
			AllowedFields: map[string]bool{
				"name": true, "q": true, "status": true, "region": true,
				"project_id": true, "size": true, "created_at_api": true,
				"first_collected_at": true, "collected_at": true,
			},
			NewQuery: func() lh.QueryAdapter {
				q := entClient.BronzeGreenNodeVolumeBlockVolume.Query()
				return lh.QueryAdapter{
					Where:      func(ps ...lh.Predicate) { q.Where(lh.ConvertSlice[predicate.BronzeGreenNodeVolumeBlockVolume](ps)...) },
					CloneCount: func(ctx context.Context) (int, error) { return q.Clone().Count(ctx) },
					Order:      func(os ...lh.Predicate) { q.Order(lh.ConvertSlice[pvol.OrderOption](os)...) },
					Fetch:      func(ctx context.Context, off, lim int) (any, error) { return q.Offset(off).Limit(lim).All(ctx) },
				}
			},
			Filters: []lh.FilterDef{
				{Field: "q", Kind: lh.Search, Pred: lh.Pred(pvol.NameContainsFold)},
				{Field: "status", Kind: lh.Multi, InFn: lh.PredIn(pvol.StatusIn), EqFn: lh.Pred(pvol.StatusEQ)},
				{Field: "region", Kind: lh.Multi, InFn: lh.PredIn(pvol.RegionIn), EqFn: lh.Pred(pvol.RegionEQ)},
				{Field: "project_id", Kind: lh.Exact, Pred: lh.Pred(pvol.ProjectIDEQ)},
			},
			SortFields: map[string]lh.SortFunc{
				"name":               lh.Sort(pvol.ByName),
				"status":             lh.Sort(pvol.ByStatus),
				"size":               lh.Sort(pvol.BySize),
				"region":             lh.Sort(pvol.ByRegion),
				"project_id":         lh.Sort(pvol.ByProjectID),
				"created_at_api":     lh.Sort(pvol.ByCreatedAtAPI),
				"collected_at":       lh.Sort(pvol.ByCollectedAt),
				"first_collected_at": lh.Sort(pvol.ByFirstCollectedAt),
			},
			DefaultOrder:  pvol.ByCollectedAt(entsql.OrderDesc()),
			FilterOptions: filterOpts,
		}),
	})

	lh.RegisterSQL(db, sqlTables)
}

func bronzeGNVolume(api, table, label string, columns []string, filters []lh.SQLFilterDef, filterOptionCols []string) lh.SQLTable {
	return lh.SQLTable{
		API:                 "/api/v1/bronze/greennode/volume/" + api,
		Schema:              "bronze",
		Table:               table,
		Nav:                 admin.NavMeta{Label: label, Group: []string{"Bronze", "GreenNode", "Volume"}},
		Columns:             columns,
		Filters:             filters,
		DefaultSort:         "collected_at",
		DefaultDesc:         true,
		FilterOptionColumns: filterOptionCols,
	}
}

var sqlTables = []lh.SQLTable{
	bronzeGNVolume("volume-types", "greennode_volume_volume_types", "Volume Types",
		[]string{"resource_id", "name", "iops", "max_size", "min_size", "through_put", "zone_id", "region", "project_id", "collected_at", "first_collected_at"},
		[]lh.SQLFilterDef{
			{Column: "name", Kind: lh.Search},
			{Column: "region", Kind: lh.Multi},
			{Column: "project_id", Kind: lh.Multi},
		},
		[]string{"region", "project_id"},
	),
	{
		API:    "/api/v1/bronze/greennode/volume/snapshots",
		Schema: "bronze",
		Table:  "greennode_volume_snapshots",
		Nav:    admin.NavMeta{Label: "Snapshots", Group: []string{"Bronze", "GreenNode", "Volume"}},
		Columns: []string{"id", "snapshot_id", "name", "size", "volume_size", "status", "created_at_api"},
		Filters: []lh.SQLFilterDef{
			{Column: "name", Kind: lh.Search},
			{Column: "status", Kind: lh.Multi},
		},
		DefaultSort:         "id",
		DefaultDesc:         true,
		FilterOptionColumns: []string{"status"},
	},
}
