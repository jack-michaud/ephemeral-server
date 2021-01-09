package bot

var AWS_VALID_REGIONS = []string {
  "us-east-1",
  "us-west-2",
}

// list of Size slug, hourly cost
const AWS_SIZE_SOURCE = "https://www.ec2instances.info/"
var AWS_VALID_SIZES = [][]string {
  {"t3a.nano", "$0.004700"},
  {"t3a.micro", "$0.009400"},
  {"t3a.small", "$0.018800"},
  {"t3a.medium", "$0.037600"},
  {"t3a.large", "$0.075200"},
  {"t3a.xlarge", "$0.150400"},
  {"t3a.2xlarge", "$0.300800"},
}
