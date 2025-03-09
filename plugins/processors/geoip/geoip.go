package geoip

import (
	"fmt"
	"net"

	"github.com/IncSW/geoip2"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

const sampleConfig = `
  ## city_db_path is the location of the MaxMind GeoIP2 City database
  city_db_path = "/var/lib/GeoIP/GeoLite2-City.mmdb"
  ## country_db_path is the location of the MaxMind GeoIP2 Country database
  # country_db_path = "/var/lib/GeoIP/GeoLite2-Country.mmdb"
  ## asn_db_path is the location of the MaxMind GeoIP2 ASN database
  # asn_db_path = "/var/lib/GeoIP/GeoLite2-ASN.mmdb"

  [[processors.geoip.lookup]]
	# get the ip from the field "source_ip" and put the lookup results in the respective destination fields (if specified)
	field = "source_ip"
	dest_country = "source_country"
	dest_city = "source_city"
	dest_lat = "source_lat"
	dest_lon = "source_lon"
	# from the ASN database
	dest_autonomous_system_organization = "source_autonomous_system_organization"
	dest_autonomous_system_number = "source_autonomous_system_number"
	dest_network = "source_network"
  `

type lookupEntry struct {
	Field                            string `toml:"field"`
	DestCountry                      string `toml:"dest_country"`
	DestCity                         string `toml:"dest_city"`
	DestLat                          string `toml:"dest_lat"`
	DestLon                          string `toml:"dest_lon"`
	DestAutonomousSystemOrganization string `toml:"dest_autonomous_system_organization"`
	DestAutonomousSystemNumber       string `toml:"dest_autonomous_system_number"`
	DestNetwork                      string `toml:"dest_network"`
}

type GeoIP struct {
	CityDBPath    string          `toml:"city_db_path"`
	CountryDBPath string          `toml:"country_db_path"`
	ASNDBPath     string          `toml:"asn_db_path"`
	Lookups       []lookupEntry   `toml:"lookup"`
	Log           telegraf.Logger `toml:"-"`

	cityReader    *geoip2.CityReader
	countryReader *geoip2.CountryReader
	asnReader     *geoip2.ASNReader
}

func (g *GeoIP) SampleConfig() string {
	return sampleConfig
}

func (g *GeoIP) Description() string {
	return "GeoIP looks up geo information and ASN details using MaxMind GeoIP2 databases"
}

func (g *GeoIP) Apply(metrics ...telegraf.Metric) []telegraf.Metric {
	for _, point := range metrics {
		for _, lookup := range g.Lookups {
			if lookup.Field == "" {
				continue
			}
			value, ok := point.GetField(lookup.Field)
			if !ok {
				continue
			}
			ipStr, ok := value.(string)
			if !ok {
				g.Log.Errorf("Field %s is not a string", lookup.Field)
				continue
			}
			ip := net.ParseIP(ipStr)
			if ip == nil {
				g.Log.Errorf("Invalid IP address: %s", ipStr)
				continue
			}

			// Process City database
			if g.cityReader != nil {
				record, err := g.cityReader.Lookup(ip)
				if err != nil {
					if err.Error() != "not found" {
						g.Log.Errorf("City GeoIP lookup error: %v", err)
					}
				} else {
					if lookup.DestCountry != "" {
						point.AddField(lookup.DestCountry, record.Country.ISOCode)
					}
					if lookup.DestCity != "" {
						point.AddField(lookup.DestCity, record.City.Names["en"])
					}
					if lookup.DestLat != "" {
						point.AddField(lookup.DestLat, record.Location.Latitude)
					}
					if lookup.DestLon != "" {
						point.AddField(lookup.DestLon, record.Location.Longitude)
					}
				}
			}

			// Process Country database
			if g.countryReader != nil {
				record, err := g.countryReader.Lookup(ip)
				if err != nil {
					if err.Error() != "not found" {
						g.Log.Errorf("Country GeoIP lookup error: %v", err)
					}
				} else {
					if lookup.DestCountry != "" {
						point.AddField(lookup.DestCountry, record.Country.ISOCode)
					}
				}
			}

			// Process ASN database
			if g.asnReader != nil {
				record, err := g.asnReader.Lookup(ip)
				if err != nil {
					if err.Error() != "not found" {
						g.Log.Errorf("ASN GeoIP lookup error: %v", err)
					}
				} else {
					if lookup.DestAutonomousSystemNumber != "" {
						point.AddField(lookup.DestAutonomousSystemNumber, record.AutonomousSystemNumber)
					}
					if lookup.DestAutonomousSystemOrganization != "" {
						point.AddField(lookup.DestAutonomousSystemOrganization, record.AutonomousSystemOrganization)
					}
					if lookup.DestNetwork != "" {
						point.AddField(lookup.DestNetwork, record.Network)
					}
				}
			}
		}
	}
	return metrics
}

func (g *GeoIP) Init() error {
	if g.CityDBPath == "" && g.CountryDBPath == "" && g.ASNDBPath == "" {
		return fmt.Errorf("at least one of city_db_path, country_db_path, or asn_db_path must be specified")
	}

	var err error
	if g.CityDBPath != "" {
		g.cityReader, err = geoip2.NewCityReaderFromFile(g.CityDBPath)
		if err != nil {
			return fmt.Errorf("error opening City GeoIP database: %v", err)
		}
	}
	if g.CountryDBPath != "" {
		g.countryReader, err = geoip2.NewCountryReaderFromFile(g.CountryDBPath)
		if err != nil {
			return fmt.Errorf("error opening Country GeoIP database: %v", err)
		}
	}
	if g.ASNDBPath != "" {
		g.asnReader, err = geoip2.NewASNReaderFromFile(g.ASNDBPath)
		if err != nil {
			return fmt.Errorf("error opening ASN GeoIP database: %v", err)
		}
	}
	return nil
}

func init() {
	processors.Add("geoip", func() telegraf.Processor {
		return &GeoIP{}
	})
}
