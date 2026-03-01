package detector

// DNSCheckDomains — small curated list of non-CDN domains for DNS integrity check.
// These have stable IPs that won't vary between resolvers (not behind Cloudflare/Akamai/etc).
var DNSCheckDomains = []string{
	"rutor.info",
	"ej.ru",
	"flibusta.is",
	"clubtone.do.am",
	"rezka.ag",
	"shikimori.one",
}

// CheckDomains — full list of domains for TLS/HTTP accessibility testing.
// Sourced from dpi-detector's domains.txt.
var CheckDomains = []string{
	"www.instagram.com",
	"www.facebook.com",
	"x.com",
	"www.linkedin.com",
	"discord.com",
	"gateway.discord.gg",
	"media.discordapp.net",
	"www.youtube.com",
	"soundcloud.com",
	"meduza.io",
	"www.dw.com",
	"www.svoboda.org",
	"holod.media",
	"www.euronews.com",
	"www.torproject.org",
	"proton.me",
	"protonvpn.com",
	"roskomsvoboda.org",
	"amnezia.org",
	"rutracker.org",
	"nnmclub.to",
	"rezka.ag",
	"hub.docker.com",
	"www.intel.com",
	"www.canva.com",
	"www.coursera.org",
	"aws.amazon.com",
	"www.apkmirror.com",
	"www.currenttime.tv",
	"www.messenger.com",
	"www.themoscowtimes.com",
	"www.linuxserver.io",
	"danbooru.donmai.us",
	"gelbooru.com",
	"www.browserleaks.com",
	"www.cdn77.com",
	"shikimori.one",
	"jut.su",
	"mos-gorsud.co",
	"kinopub.online",
	"kino.pub",
}

// DNS servers for UDP resolution comparison
var UDPDNSServers = []string{
	"8.8.8.8",         // Google
	"1.1.1.1",         // Cloudflare
	"9.9.9.9",         // Quad9
	"94.140.14.14",    // AdGuard
	"77.88.8.8",       // Yandex
	"223.5.5.5",       // Alibaba
	"208.67.222.222",  // OpenDNS
	"76.76.2.0",       // ControlD
	"194.242.2.2",     // Mullvad
}

// DoH (DNS-over-HTTPS) endpoints
var DoHServers = []struct {
	Name string
	URL  string
}{
	{"Google (IP)", "https://8.8.8.8/resolve"},
	{"Google", "https://dns.google/resolve"},
	{"Cloudflare (IP)", "https://1.1.1.1/dns-query"},
	{"Cloudflare", "https://cloudflare-dns.com/dns-query"},
	{"AdGuard", "https://dns.adguard-dns.com/resolve"},
}

// Block page markers in URL redirects
var BlockMarkers = []string{
	"lawfilter",
	"warning.rt.ru",
	"blocked",
	"access-denied",
	"eais",
	"zapret-info",
	"rkn.gov.ru",
	"mvd.ru",
}

// Block page markers in HTTP response body
var BodyBlockMarkers = []string{
	"blocked",
	"заблокирован",
	"запрещён",
	"запрещен",
	"ограничен",
	"единый реестр",
	"роскомнадзор",
	"rkn.gov.ru",
	"nap.gov.ru",
	"eais.rkn.gov.ru",
	"warning.rt.ru",
	"blocklist",
	"решению суда",
}

// TCPTargets — CDN/hosting endpoints for TCP 16-20KB drop test.
// Sourced from dpi-detector's tcp_16_20_targets.json.
var TCPTargets = []TCPTarget{
	{"SE.AKM-01", "https://media.miele.com/images/2000015/200001503/20000150334.png", "AS20940", "Akamai", "SE"},
	{"US.AKM-02", "https://www.roxio.com/static/roxio/videos/products/nxt9/lamp-magic.mp4", "AS16625", "Akamai", "US"},
	{"US.AKM-03", "http://speedtest.newark.linode.com/100MB-newark.bin", "AS63949", "Akamai HTTP", "US"},
	{"FR.AKM-04", "https://www.rbcroyalbank.com/dvl/v1.0/assets/fonts/Roboto-Light.woff", "AS16625", "Akamai", "FR"},
	{"FR.AKM-05", "https://www.thomascook.in/js/updatedHomeLib.js?version=1.5", "AS16625", "Akamai", "FR"},
	{"DE.AWS-01", "https://corp.kaltura.com/wp-content/cache/min/1/wp-content/themes/airfleet/dist/styles/theme.css", "AS16509", "AWS", "DE"},
	{"FR.AWS-02", "https://www.herokucdn.com/malibu/latest/sprite.svg", "AS16509", "AWS", "FR"},
	{"DE.AWS-03", "https://www.getscope.com/assets/fonts/fa-solid-900.woff2", "AS16509", "AWS", "DE"},
	{"GB.AWS-04", "https://www.zillowstatic.com/s3/constellation-website/public/shared/fonts/open-sans/LATEST/open-sans-variable.woff2", "AS16509", "AWS", "GB"},
	{"FR.C77-01", "https://cdn.eso.org/images/banner1920/eso2520a.jpg", "AS60068", "CDN77", "FR"},
	{"FR.C77-02", "https://i8secure14-805356.c.cdn77.org/mm/Customers_File/website/bgimages/6c9e8c8c-e7c5-45b2-a9d8-b3fdfed4177d/slide_20250629_MMS_SS25-26-3831_C_1920x1080px_SFW.jpg", "AS60068", "CDN77", "FR"},
	{"CA.CF-01", "https://aegis.audioeye.com/assets/index.js", "AS13335", "Cloudflare", "CA"},
	{"US.CF-02", "https://esm.sh/gh/esm-dev/esm.sh@e7447dea04/server/embed/assets/sceenshot-deno-types.png", "AS13335", "Cloudflare", "US"},
	{"CA.CF-03", "https://img.wzstats.gg/cleaver/gunFullDisplay", "AS13335", "Cloudflare", "CA"},
	{"US.CF-04", "https://www.bigcartel.com/_next/image?url=https%3A%2F%2Fimages.prismic.io%2Fbigcartel-staging%2FaAkmrfIqRLdaBiNZ_home_hero_lifestyle.png%3Fauto%3Dformat%2Ccompress%26rect%3D0%2C0%2C1600%2C1600%26w%3D1200%26h%3D1200&w=3840&q=75", "AS13335", "Cloudflare", "US"},
	{"CA.CF-05", "https://www.labcorp.com/content/dam/labcorp/videos/homepage/testfinder-card.mp4", "AS13335", "Cloudflare", "CA"},
	{"CA.CF-06", "https://static.generated.photos/vue-static/home/solutions/humans.webp", "AS13335", "Cloudflare", "CA"},
	{"US.CNST-01", "https://static-cdn.play.date/static/js/model-viewer.min.js", "AS20473", "Constant", "US"},
	{"NL.CNST-02", "https://viaanabel.al/static/banners/banner_1001_v3_sq.jpg", "AS20473", "Constant", "NL"},
	{"CL.CNST-03", "https://ctcu.com.ar/download/novedades.imagen.886cc162ee9fa24a.576861747341707020496d61676520323032352d31322d31312061742030382e35322e33362e6a706567.jpeg?_signature=liCl4x7kRPl4p8Tk4ivx3p-82Ig", "AS20473", "Constant", "CL"},
	{"FR.CNTB-01", "https://findair.net/wp-content/uploads/2025/07/online-booking-2.jpeg", "AS51167", "Contabo", "FR"},
	{"FR.CNTB-02", "https://programmer.am/css/style.css", "AS51167", "Contabo", "FR"},
	{"FR.CNTB-03", "https://nare.am/wp-content/uploads/2020/06/nare_armenia_travel-1280x580.jpg", "AS51167", "Contabo", "FR"},
	{"FR.CNTB-04", "https://metropolis.al/wp-content/uploads/2023/12/mt-sample-background.jpg", "AS51167", "Contabo", "FR"},
	{"US.DO-01", "https://ecomstal.com/_next/static/css/73cc557714b4846b.css", "AS14061", "DigitalOcean", "US"},
	{"US.DO-02", "https://opennetworking.org/wp-content/themes/onf/main.css", "AS14061", "DigitalOcean", "US"},
	{"US.DO-03", "https://carishealthcare.com/content/uploads/2025/04/Rectangle-105.jpg", "AS14061", "DigitalOcean", "US"},
	{"US.DO-04", "https://bohnlawllc.com/wp-content/uploads/sites/27/2024/01/Trusts.jpg", "AS14061", "DigitalOcean", "US"},
	{"GB.DO-05", "https://www.linuxserver.io/user/pages/01.home/03._03_standard/rsz_fancycrave-151127-unsplash.jpg", "AS14061", "DigitalOcean", "GB"},
	{"CA.FST-01", "https://www.jetblue.com/footer/footer-element-es2015.js", "AS54113", "Fastly", "CA"},
	{"CA.FST-02", "https://ssl.p.jwpcdn.com/player/v/8.40.5/bidding.js", "AS54113", "Fastly", "CA"},
	{"LU.GCORE-01", "https://gcore.com/assets/fonts/Montserrat-Variable.woff2", "AS199524", "Gcore", "LU"},
	{"US.GC-01", "https://api.usercentrics.eu/gvl/v3/en.json", "AS396982", "Google Cloud", "US"},
	{"US.GC-02", "https://cromwell-intl.com/fonts/hammersmithone.ttf", "AS396982", "Google Cloud", "US"},
	{"DE.HE-01", "https://apiwhatsapp-1000.zapipro.com/libs/bootstrap/dist/css/bootstrap.min.css", "AS24940", "Hetzner", "DE"},
	{"DE.HE-02", "https://www.industrialport.net/wp-content/uploads/custom-fonts/2022/10/Lato-Bold.ttf", "AS24940", "Hetzner", "DE"},
	{"FI.HE-04", "https://251b5cd9.nip.io/1MB.bin", "AS24940", "Hetzner", "FI"},
	{"FI.HE-05", "https://nioges.com/libs/fontawesome/webfonts/fa-solid-900.woff2", "AS24940", "Hetzner", "FI"},
	{"FI.HE-06", "https://5fd8bdae.nip.io/1MB.bin", "AS24940", "Hetzner", "FI"},
	{"FI.HE-07", "https://5fd8bca5.nip.io/1MB.bin", "AS24940", "Hetzner", "FI"},
	{"DE.HE-01H", "http://media5.cdnbase.com/media/photologue/photos/6143813.jpg", "AS24940", "Hetzner HTTP", "DE"},
	{"NL.LSW-01", "https://mirror.leaseweb.com/alpine/v3.9/releases/x86_64/alpine-extended-3.9.0-x86_64.iso", "AS60781", "Leaseweb", "NL"},
	{"US.MBC-01", "https://twin.mentat.su/assets/fonts/Inter-SemiBold.woff2", "AS8849", "Melbicom", "US"},
	{"MX.OR-01", "http://40.233.0.95/assets/bundle.538a44e1.js", "AS31898", "Oracle HTTP", "MX"},
	{"MX.OR-02", "https://k.860617.xyz/static/app/dist/main.js", "AS31898", "Oracle", "MX"},
	{"SG.OR-03", "https://global-seres.com.sg/wp-content/uploads/2024/02/SVG00732-scaled.jpg", "AS31898", "Oracle", "SG"},
	{"SG.OR-04", "https://www.citrusmedia.com.sg/wp-content/plugins/elementor/assets/lib/font-awesome/webfonts/fa-brands-400.woff2", "AS31898", "Oracle", "SG"},
	{"CO.OR-05", "https://plataforma.trackerintl.com/images/background.jpg", "AS31898", "Oracle", "CO"},
	{"FR.OVH-01", "https://proof.ovh.net/files/1Mb.dat", "AS16276", "OVH", "FR"},
	{"FR.OVH-02", "https://proof.ovh.net/files/10Mb.dat", "AS16276", "OVH", "FR"},
	{"FR.OVH-03", "https://app.symarobot.com/content/images/logo.png", "AS16276", "OVH", "FR"},
	{"CA.OVH-04", "https://proof.ovh.ca/files/100Mb.dat", "AS16276", "OVH", "CA"},
	{"FR.OVH-05", "https://filmoteka.net.pl/css/bootstrap.min.css", "AS16276", "OVH", "FR"},
	{"NL.SW-01", "https://www.velivole.fr/img/header.jpg", "AS12876", "Scaleway", "NL"},
	{"FR.SW-02", "https://www.moobicom.ci/assets/slider1.jpg", "AS12876", "Scaleway", "FR"},
	{"FR.SW-03", "https://www.zenetys.com/en/", "AS12876", "Scaleway", "FR"},
	{"FR.SW-04", "https://www.logvault.io/assets/Poppins-Regular-CTKNfV9P.ttf", "AS12876", "Scaleway", "FR"},
	{"DE.VLTR-01", "https://static-cdn.play.date/static/js/model-viewer.min.js", "AS20473", "Vultr", "DE"},
	{"US.VLTR-02", "https://us.rudder.qntmnet.com/QN-CDN/images/qn_bg_.jpg", "AS20473", "Vultr", "US"},
	{"DE.HOST-01", "https://kast-tv.ru/fonts/GraphikLCGRegular.woff", "AS216127", "nuxt.cloud", "DE"},
	{"MD.HOST-02", "https://profinance.cc/img/landing/introduction.png", "AS200019", "Alexhost", "MD"},
	{"FI.HOST-03", "https://cascademl.com/images/5.jpg", "AS215730", "H2nexus", "FI"},
}
