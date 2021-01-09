package bot

var DIGITALOCEAN_VALID_REGIONS = []string {
  "nyc1",
  "sfo1",
  "nyc2",
  "ams2",
  "sgp1",
  "lon1",
  "nyc3",
  "ams3",
  "fra1",
  "tor1",
  "sfo2",
  "blr1",
  "sfo3",
}

// list of  Size slug, hourly cost
const DIGITALOCEAN_SIZE_SOURCE = "https://slugs.do-api.dev/"
var DIGITALOCEAN_VALID_SIZES = [][]string {
  {"s-8vcpu-16gb", "$0.11905"},
  {"s-4vcpu-8gb", "$0.05952"},
  {"s-2vcpu-4gb", "$0.02976"},
  {"s-2vcpu-2gb", "$0.02232"},
  {"s-1vcpu-2gb", "$0.01488"},
  {"s-1vcpu-1gb", "$0.00744"},
}
