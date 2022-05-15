package common

import (
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/parquet"
	"github.com/xitongsys/parquet-go/reader"
	"github.com/xitongsys/parquet-go/writer"
)

const (
	CommonStock string = "Common Stock"
	ETF         string = "Exchange Traded Fund"
	ETN         string = "Exchange Traded Note"
	Fund        string = "Closed-End Fund"
	MutualFund  string = "Mutual Fund"
	ADRC        string = "American Depository Receipt Common"
)

type Asset struct {
	Ticker               string   `json:"ticker" parquet:"name=ticker, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Name                 string   `json:"Name" parquet:"name=name, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Description          string   `json:"description" parquet:"name=description, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	PrimaryExchange      string   `json:"primary_exchange" parquet:"name=primary_exchange, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	AssetType            string   `json:"asset_type" parquet:"name=asset_type, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	CompositeFigi        string   `json:"composite_figi" parquet:"name=composite_figi, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	ShareClassFigi       string   `json:"share_class_figi" parquet:"name=share_class_figi, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	CUSIP                string   `json:"cusip" parquet:"name=cusip, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	ISIN                 string   `json:"isin" parquet:"name=isin, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	CIK                  string   `json:"cik" parquet:"name=cik, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	ListingDate          string   `json:"listing_date" parquet:"name=listing_date, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	DelistingDate        string   `json:"delisting_date" parquet:"name=delisting_date, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Industry             string   `json:"industry" parquet:"name=industry, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Sector               string   `json:"sector" parquet:"name=sector, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Icon                 []byte   `json:"icon"`
	IconUrl              string   `json:"icon_url" parquet:"name=icon_url, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	CorporateUrl         string   `json:"corporate_url" parquet:"name=corporate_url, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	HeadquartersLocation string   `json:"headquarters_location" parquet:"name=headquarters_location, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	SimilarTickers       []string `json:"similar_tickers" parquet:"name=similar_tickers, type=MAP, convertedtype=LIST, valuetype=BYTE_ARRAY, valueconvertedtype=UTF8"`
	PolygonDetailAge     int64    `json:"polygon_detail_age" parquet:"name=polygon_detail_age, type=INT64"`
	LastUpdated          int64    `json:"last_updated" parquet:"name=last_update, type=INT64"`
	Source               string   `json:"source" parquet:"name=source, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
}

func TrimWhiteSpace(assets []*Asset) {
	for _, asset := range assets {
		asset.Name = strings.TrimSpace(asset.Name)
		asset.Description = strings.TrimSpace(asset.Description)
		asset.CIK = strings.TrimSpace(asset.CIK)
		asset.CUSIP = strings.TrimSpace(asset.CUSIP)
		asset.Industry = strings.TrimSpace(asset.Industry)
		asset.Sector = strings.TrimSpace(asset.Sector)
		asset.ISIN = strings.TrimSpace(asset.ISIN)
	}
}

func ReadFromParquet(fn string) []*Asset {
	log.Info().Str("FileName", fn).Msg("loading parquet file")
	fr, err := local.NewLocalFileReader(fn)
	if err != nil {
		log.Error().Err(err).Msg("can't open file")
		return nil
	}

	pr, err := reader.NewParquetReader(fr, new(Asset), 4)
	if err != nil {
		log.Error().Err(err).Msg("can't create parquet reader")
		return nil
	}

	num := int(pr.GetNumRows())
	rec := make([]*Asset, num)
	if err = pr.Read(&rec); err != nil {
		log.Error().Err(err).Msg("parquet read error")
		return nil
	}

	pr.ReadStop()
	fr.Close()

	return rec
}

func SaveToParquet(records []*Asset, fn string) error {
	var err error

	fh, err := local.NewLocalFileWriter(fn)
	if err != nil {
		log.Error().Err(err).Str("FileName", fn).Msg("cannot create local file")
		return err
	}
	defer fh.Close()

	pw, err := writer.NewParquetWriter(fh, new(Asset), 4)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Parquet write failed")
		return err
	}

	pw.RowGroupSize = 128 * 1024 * 1024 // 128M
	pw.PageSize = 8 * 1024              // 8k
	pw.CompressionType = parquet.CompressionCodec_GZIP

	for _, r := range records {
		if r.DelistingDate != "" {
			continue
		}
		if err = pw.Write(r); err != nil {
			log.Error().
				Err(err).
				Str("CompositeFigi", r.CompositeFigi).
				Msg("Parquet write failed for record")
		}
	}

	if err = pw.WriteStop(); err != nil {
		log.Error().Err(err).Msg("Parquet write failed")
		return err
	}

	log.Info().Int("NumRecords", len(records)).Msg("parquet write finished")
	return nil
}

func (asset *Asset) MarshalZerologObject(e *zerolog.Event) {
	e.Str("Ticker", asset.Ticker)
	e.Str("Name", asset.Name)
	e.Str("Description", asset.Description)
	e.Str("PrimaryExchange", asset.PrimaryExchange)
	e.Str("AssetType", string(asset.AssetType))
	e.Str("CompositeFigi", asset.CompositeFigi)
	e.Str("ShareClassFigi", asset.ShareClassFigi)
	e.Str("CUSIP", asset.CUSIP)
	e.Str("ISIN", asset.ISIN)
	e.Str("CIK", asset.CIK)
	e.Str("ListingDate", asset.ListingDate)
	e.Str("DelistingDate", asset.DelistingDate)
	e.Str("Industry", asset.Industry)
	e.Str("Sector", asset.Sector)
	e.Str("IconUrl", asset.IconUrl)
	e.Str("CorporateUrl", asset.CorporateUrl)
	e.Str("HeadquartersLocation", asset.HeadquartersLocation)
	e.Str("Source", asset.Source)
	e.Int64("PolygonDetailAge", asset.PolygonDetailAge)
	e.Int64("LastUpdate", asset.LastUpdated)
}
