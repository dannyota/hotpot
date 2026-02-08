env "bronze" {
  src = "ent://../../pkg/storage/ent/bronze/atlas_schema"
  dev = env("HOTPOT_DEV_DATABASE_URL") # Read from env var (set by migrate tool)
  url = env("HOTPOT_DATABASE_URL")     # Read from env var (set by migrate tool)
  migration {
    dir = "file://bronze"
  }
}

env "bronzehistory" {
  src = "ent://../../pkg/storage/ent/bronzehistory/atlas_schema"
  dev = env("HOTPOT_DEV_DATABASE_URL") # Read from env var (set by migrate tool)
  url = env("HOTPOT_DATABASE_URL")     # Read from env var (set by migrate tool)
  migration {
    dir = "file://bronzehistory"
  }
}

env "silver" {
  src = "ent://../../pkg/storage/ent/silver/atlas_schema"
  dev = env("HOTPOT_DEV_DATABASE_URL") # Read from env var (set by migrate tool)
  url = env("HOTPOT_DATABASE_URL")     # Read from env var (set by migrate tool)
  migration {
    dir = "file://silver"
  }
}

env "gold" {
  src = "ent://../../pkg/storage/ent/gold/atlas_schema"
  dev = env("HOTPOT_DEV_DATABASE_URL") # Read from env var (set by migrate tool)
  url = env("HOTPOT_DATABASE_URL")     # Read from env var (set by migrate tool)
  migration {
    dir = "file://gold"
  }
}
