package commands

import (
	"fmt"
	"os"
	"strings"
	"encoding/json"

	"github.com/activecm/rita/pkg/beacon"
	"github.com/activecm/rita/resources"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
)

func init() {
	command := cli.Command{
		Name:      "show-beacons",
		Usage:     "Print hosts which show signs of C2 software",
		ArgsUsage: "<database>",
		Flags: []cli.Flag{
			humanFlag,
			configFlag,
			delimFlag,
			netNamesFlag,
			jsonOutputFlag,
		},
		Action: showBeacons,
	}

	bootstrapCommands(command)
}

func showBeacons(c *cli.Context) error {
	db := c.Args().Get(0)
	if db == "" {
		return cli.NewExitError("Specify a database", -1)
	}
	res := resources.InitResources(c.String("config"))
	res.DB.SelectDB(db)

	data, err := beacon.Results(res, 0)

	if err != nil {
		res.Log.Error(err)
		return cli.NewExitError(err, -1)
	}

	if !(len(data) > 0) {
		return cli.NewExitError("No results were found for "+db, -1)
	}

	showNetNames := c.Bool("network-names")

	if c.Bool("json-output") {
		showBeaconsJSON(data, showNetNames)
		return nil
	}
	
	if c.Bool("human-readable") {
		err := showBeaconsHuman(data, showNetNames)
		if err != nil {
			return cli.NewExitError(err.Error(), -1)
		}
		return nil
	}

	err = showBeaconsDelim(data, c.String("delimiter"), showNetNames)
	if err != nil {
		return cli.NewExitError(err.Error(), -1)
	}
	return nil
}

func showBeaconsHuman(data []beacon.Result, showNetNames bool) error {
	table := tablewriter.NewWriter(os.Stdout)
	var headerFields []string
	if showNetNames {
		headerFields = []string{
			"Score", "Source Network", "Destination Network", "Source IP", "Destination IP",
			"Connections", "Avg. Bytes", "Intvl Range", "Size Range", "Top Intvl",
			"Top Size", "Top Intvl Count", "Top Size Count", "Intvl Skew",
			"Size Skew", "Intvl Dispersion", "Size Dispersion",
		}
	} else {
		headerFields = []string{
			"Score", "Source IP", "Destination IP",
			"Connections", "Avg. Bytes", "Intvl Range", "Size Range", "Top Intvl",
			"Top Size", "Top Intvl Count", "Top Size Count", "Intvl Skew",
			"Size Skew", "Intvl Dispersion", "Size Dispersion",
		}
	}

	table.SetHeader(headerFields)

	for _, d := range data {
		var row []string
		if showNetNames {
			row = []string{
				f(d.Score), d.SrcNetworkName, d.DstNetworkName,
				d.SrcIP, d.DstIP, i(d.Connections), f(d.AvgBytes),
				i(d.Ts.Range), i(d.Ds.Range), i(d.Ts.Mode), i(d.Ds.Mode),
				i(d.Ts.ModeCount), i(d.Ds.ModeCount), f(d.Ts.Skew), f(d.Ds.Skew),
				i(d.Ts.Dispersion), i(d.Ds.Dispersion),
			}
		} else {
			row = []string{
				f(d.Score), d.SrcIP, d.DstIP, i(d.Connections), f(d.AvgBytes),
				i(d.Ts.Range), i(d.Ds.Range), i(d.Ts.Mode), i(d.Ds.Mode),
				i(d.Ts.ModeCount), i(d.Ds.ModeCount), f(d.Ts.Skew), f(d.Ds.Skew),
				i(d.Ts.Dispersion), i(d.Ds.Dispersion),
			}
		}
		table.Append(row)
	}
	table.Render()
	return nil
}

func showBeaconsDelim(data []beacon.Result, delim string, showNetNames bool) error {
	var headerFields []string
	if showNetNames {
		headerFields = []string{
			"Score", "Source Network", "Destination Network", "Source IP", "Destination IP",
			"Connections", "Avg. Bytes", "Intvl Range", "Size Range", "Top Intvl",
			"Top Size", "Top Intvl Count", "Top Size Count", "Intvl Skew",
			"Size Skew", "Intvl Dispersion", "Size Dispersion",
		}
	} else {
		headerFields = []string{
			"Score", "Source IP", "Destination IP",
			"Connections", "Avg. Bytes", "Intvl Range", "Size Range", "Top Intvl",
			"Top Size", "Top Intvl Count", "Top Size Count", "Intvl Skew",
			"Size Skew", "Intvl Dispersion", "Size Dispersion",
		}
	}

	// Print the headers and analytic values, separated by a delimiter
	fmt.Println(strings.Join(headerFields, delim))
	for _, d := range data {

		var row []string
		if showNetNames {
			row = []string{
				f(d.Score), d.SrcNetworkName, d.DstNetworkName,
				d.SrcIP, d.DstIP, i(d.Connections), f(d.AvgBytes),
				i(d.Ts.Range), i(d.Ds.Range), i(d.Ts.Mode), i(d.Ds.Mode),
				i(d.Ts.ModeCount), i(d.Ds.ModeCount), f(d.Ts.Skew), f(d.Ds.Skew),
				i(d.Ts.Dispersion), i(d.Ds.Dispersion),
			}
		} else {
			row = []string{
				f(d.Score), d.SrcIP, d.DstIP, i(d.Connections), f(d.AvgBytes),
				i(d.Ts.Range), i(d.Ds.Range), i(d.Ts.Mode), i(d.Ds.Mode),
				i(d.Ts.ModeCount), i(d.Ds.ModeCount), f(d.Ts.Skew), f(d.Ds.Skew),
				i(d.Ts.Dispersion), i(d.Ds.Dispersion),
			}
		}

		fmt.Println(strings.Join(row, delim))
	}
	return nil
}

func showBeaconsJSON(data []beacon.Result, showNetNames bool) error {
	type DataRow struct {
		Score float64
		SourceNetworkName string `json:",omitempty"`
		DestinationNetworkName string	`json:",omitempty"`
		SourceIP string
		DestinationIP string
		Connections int64
		AvgBytes float64
		IntvlRange int64
		SizeRange int64
		TopIntvl int64
		TopSize int64
		TopIntvlCount int64
		TopSizeCount int64
		IntvlSkew float64
		SizeSkew float64
		IntvlDispersion int64
		SizeDispersion int64
	}

	rows := make([]DataRow, len(data)-1)
	for _, d := range data {
		srcNetworkName := ""
		dstNetworkName := ""
		if showNetNames {
			srcNetworkName = d.SrcNetworkName
			dstNetworkName = d.DstNetworkName
		}

		row := DataRow{d.Score, srcNetworkName, dstNetworkName, d.SrcIP, d.DstIP, d.Connections, d.AvgBytes,
			d.Ts.Range, d.Ds.Range, d.Ts.Mode, d.Ds.Mode, d.Ts.ModeCount,
			d.Ds.ModeCount, d.Ts.Skew, d.Ds.Skew, d.Ts.Dispersion, d.Ds.Dispersion}
		rows = append(rows, row)
	}

	rowsJson, err := json.Marshal(rows)
	if err != nil {
		return cli.NewExitError(err.Error(), -1)
	}

	jsonOutput := string(rowsJson[:])
	fmt.Println(jsonOutput)

	return nil
}
