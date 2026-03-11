package config

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"
)

type hostingEntry struct {
	value       string
	description string
	country     string // "" = NULL
}

// SeedHostingIndicators inserts predefined hosting domain and keyword indicators.
func SeedHostingIndicators(ctx context.Context, db *sql.DB) error {
	now := time.Now()

	// Domains.
	if err := seedHostingBatch(ctx, db, "domain", hostingDomains, now); err != nil {
		return err
	}
	// Keywords.
	return seedHostingBatch(ctx, db, "keyword", hostingKeywords, now)
}

func seedHostingBatch(ctx context.Context, db *sql.DB, indicatorType string, entries []hostingEntry, now time.Time) error {
	if len(entries) == 0 {
		return nil
	}

	var b strings.Builder
	b.WriteString(`INSERT INTO config.hosting_indicators
		(indicator_type, value, description, country, is_active, created_at, updated_at)
		VALUES `)

	args := make([]any, 0, len(entries)*7)
	for i, e := range entries {
		if i > 0 {
			b.WriteString(", ")
		}
		base := i * 7
		fmt.Fprintf(&b, "($%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			base+1, base+2, base+3, base+4, base+5, base+6, base+7)

		var country *string
		if e.country != "" {
			country = &e.country
		}
		args = append(args, indicatorType, e.value, e.description, country, true, now, now)
	}

	b.WriteString(` ON CONFLICT (indicator_type, value) DO NOTHING`)

	_, err := db.ExecContext(ctx, b.String(), args...)
	if err != nil {
		return fmt.Errorf("upsert hosting indicators (%s, %d entries): %w", indicatorType, len(entries), err)
	}
	return nil
}

var hostingDomains = func() []hostingEntry {
	entries := []hostingEntry{
		// Global hyperscale cloud
		{"amazon.com", "AWS", ""},
		{"google.com", "Google Cloud", ""},
		{"microsoft.com", "Azure", ""},
		{"oracle.com", "Oracle Cloud", ""},
		{"ibm.com", "IBM Cloud", ""},
		{"alibaba.com", "Alibaba Cloud", ""},
		{"alibabagroup.com", "Alibaba Cloud (alternate domain)", ""},
		{"tencent.com", "Tencent Cloud", "CN"},
		{"apple.com", "Apple (iCloud Private Relay)", ""},

		// Global hosting / CDN / datacenter
		{"akamai.com", "Akamai CDN", ""},
		{"linode.com", "Linode/Akamai", ""},
		{"digitalocean.com", "DigitalOcean", ""},
		{"vultr.com", "Vultr", ""},
		{"equinix.com", "Equinix Metal", ""},
		{"rackspace.com", "Rackspace", ""},
		{"leaseweb.com", "LeaseWeb", ""},
		{"online.net", "Scaleway", ""},
		{"scaleway.com", "Scaleway", ""},
		{"zscaler.com", "Zscaler", ""},
		{"he.net", "Hurricane Electric", ""},
		{"constant.com", "The Constant Company", ""},
		{"hivelocity.net", "Hivelocity", ""},
		{"zenlayer.com", "Zenlayer", ""},
		{"latitude.sh", "Latitude.sh (bare metal)", ""},
		{"clouvider.com", "Clouvider", ""},
		{"gcore.com", "G-Core Labs CDN", ""},
		{"selectel.com", "Selectel", ""},
		{"ipxo.com", "IPXO (IP leasing)", ""},
		{"m247global.com", "M247", ""},
		{"globalsecurelayer.com", "Global Secure Layer", ""},
		{"packethub.net", "PacketHub", ""},
		{"unitasglobal.com", "Unitas Global", ""},
		{"larus.net", "LARUS", ""},
		{"modmc.net", "MOD Mission Critical", ""},
		{"simoresta.lt", "Simoresta (Lithuania)", "LT"},
		{"ntt.com", "NTT (Global IP Network)", ""},
		{"global.ntt", "NTT Global", ""},
		{"datacamp.co.uk", "Datacamp Limited", "GB"},

		// Vietnam hosting / IDC
		{"cmctelecom.vn", "CMC Telecom (major VN IDC)", "VN"},
		{"fpt.com.vn", "FPT Telecom (IDC + hosting)", "VN"},
		{"fpt.vn", "FPT Telecom", "VN"},
		{"netnam.vn", "Netnam", "VN"},
		{"vietserver.vn", "VietServer", "VN"},
		{"viettelidc.com.vn", "Viettel IDC", "VN"},
		{"isvc.vn", "Viet Digital Technology", "VN"},
		{"eztech.com.vn", "EZ Technology", "VN"},
		{"vng.com.vn", "VNG Corporation", "VN"},
		{"bkns.vn", "Bach Kim Network Solutions", "VN"},
		{"hitc.vn", "Hanoi Telecom", "VN"},
		{"ods.vn", "ODS", "VN"},
		{"superdata.vn", "SuperData", "VN"},
		{"lvsolution.vn", "Long Van System Solution", "VN"},
		{"vnetwork.vn", "VNetwork", "VN"},
		{"vndata.vn", "Viet Storage Technology", "VN"},
		{"idconline.vn", "IDC Online", "VN"},
		{"gofiber.vn", "GoFiber", "VN"},
		{"tino.vn", "Tino Group", "VN"},
		{"vhost.vn", "VHost", "VN"},
		{"trumvps.vn", "TrumVPS", "VN"},
		{"123host.vn", "123Host", "VN"},
		{"techhost.vn", "TechHost", "VN"},
		{"inet.vn", "iNET", "VN"},
		{"spt.vn", "Saigon Postel", "VN"},

		// China hosting / IDC
		{"ucloud.cn", "UCloud", "CN"},
		{"sinnet.com.cn", "Beijing Sinnet (AWS China partner)", "CN"},
		{"zeofast.cn", "Zeofast Network", "CN"},

		// Japan hosting / IDC
		{"sakura.ad.jp", "Sakura Internet", "JP"},
		{"idcf.jp", "IDC Frontier", "JP"},
		{"xserver.co.jp", "Xserver", "JP"},
		{"gmointernet.jp", "GMO Internet", "JP"},

		// Korea hosting / IDC
		{"moack.co.kr", "MOACK", "KR"},
		{"abcle.co.kr", "Abcle", "KR"},
		{"piranha.co.kr", "Piranha Systems", "KR"},
		{"hostway.co.kr", "Hostway IDC", "KR"},
		{"hostcenter.co.kr", "Hostcenter", "KR"},
		{"connectwave.co.kr", "Connectwave", "KR"},
		{"flexnetworks.co.kr", "Flexnetworks", "KR"},

		// Taiwan hosting / IDC
		{"chief.com.tw", "Chief Telecom", "TW"},
		{"serverfield.com.tw", "Serverfield", "TW"},

		// Singapore hosting / IDC
		{"netstartech.sg", "Netstar", "SG"},
		{"anydigital.sg", "Any Digital", "SG"},
		{"readyserver.sg", "Ready Server", "SG"},
		{"scloud.sg", "SCloud", "SG"},
		{"telin.sg", "Telkom Indonesia International", "SG"},

		// Thailand hosting / IDC
		{"jastel.co.th", "JasTel Network", "TH"},

		// Indonesia hosting / IDC
		{"cbn.id", "Cyberindo Aditama", "ID"},
		{"iforte.co.id", "iForte", "ID"},
		{"myrepublic.co.id", "MyRepublic", "ID"},
		{"nusa.net.id", "Nusa Net", "ID"},
		{"arupa.id", "Arupa Cloud Nusantara", "ID"},
		{"wowrack.co.id", "Wowrack", "ID"},
		{"hypernet.co.id", "Hypernet", "ID"},

		// Malaysia hosting / IDC
		{"gbnetwork.my", "GB Network Solutions", "MY"},
		{"aims.com.my", "AIMS Data Centre", "MY"},
		{"lcsb.my", "Light Cloud Technology", "MY"},

		// Hong Kong hosting / IDC
		{"powerline.hk", "Power Line Datacenter", "HK"},
		{"akari.hk", "Akari Networks", "HK"},
		{"udc.hk", "UDC", "HK"},
		{"udomain.hk", "UDomain", "HK"},
		{"cloudie.hk", "Cloudie", "HK"},

		// India hosting / IDC
		{"webwerks.in", "Web Werks", "IN"},
		{"ctrls.in", "CtrlS", "IN"},
		{"pioneer.co.in", "Pioneer Elabs", "IN"},
		{"sifycorp.com", "Sify (India)", "IN"},

		// Philippines hosting
		{"eastern.com.ph", "Eastern Telecoms", "PH"},

		// Russia hosting / IDC
		{"rt.ru", "Rostelecom (includes IDC)", "RU"},
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].value < entries[j].value })
	return entries
}()

var hostingKeywords = []hostingEntry{
	{"host", "Substring match for hosting providers", ""},
	{"cloud", "Substring match for cloud providers", ""},
	{"server", "Substring match for server providers", ""},
	{"datacenter", "Substring match for datacenters", ""},
	{"data-center", "Substring match for datacenters (hyphenated)", ""},
	{"vps", "Substring match for VPS providers", ""},
	{"cdn", "Substring match for CDN providers", ""},
	{"colocation", "Substring match for colocation providers", ""},
	{"colo", "Substring match for colocation providers (short)", ""},
	{"dedicated", "Substring match for dedicated server providers", ""},
}
