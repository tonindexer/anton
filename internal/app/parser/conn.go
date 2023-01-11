package parser

type connInfo struct {
	addr string
	key  string
}

var mainnetArchive = []*connInfo{
	{addr: "135.181.177.59:53312", key: "aF91CuUHuuOv9rm2W5+O/4h38M3sRm40DtSdRxQhmtQ="},
	// {addr: "164.68.101.206:52995", key: "QnGFe9kihW+TKacEvvxFWqVXeRxCB6ChjjhNTrL7+/k="},
	// {addr: "164.68.99.144:20334", key: "gyLh12v4hBRtyBygvvbbO2HqEtgl+ojpeRJKt4gkMq0="},
	{addr: "188.68.216.239:19925", key: "ucho5bEkufbKN1JR1BGHpkObq602whJn3Q3UwhtgSo4="},
	// {addr: "54.39.158.156:51565", key: "TDg+ILLlRugRB4Kpg3wXjPcoc+d+Eeb7kuVe16CS9z8="},
}

var testnetArchive = []*connInfo{
	{addr: "65.108.141.177:17439", key: "0MIADpLH4VQn+INHfm0FxGiuZZAA8JfTujRqQugkkA8="},
	// {addr: "116.202.208.174:51281", key: "hyXd2d6yyiD/wirjoraSrKek1jYhOyzbQoIzV85CB98="},
	// {addr: "161.97.147.21:2411", key: "/qYMangDWyMRqZl4tAH1gKBxwuvD54EKKGhNgBMQ1Tk="},
	// {addr: "65.108.204.54:29296", key: "p2tSiaeSqX978BxE5zLxuTQM06WVDErf5/15QToxMYA="},
}
