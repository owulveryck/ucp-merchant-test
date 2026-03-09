package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strings"

	icatalog "github.com/owulveryck/ucp-merchant-test/internal/catalog"
)

type productTemplate struct {
	title              string
	category           string
	brand              string
	price              int
	slug               string
	description        string
	usageType          string
	availableCountries []string
}

var templates = []productTemplate{
	{"Wireless Headphones", "Electronics", "TechFlow", 7999, "headphones",
		"Lightweight on-ear headphones for casual music listening and video calls. 15-hour battery with quick charge. Perfect for commuters and weekend use.",
		"occasional", nil},
	{"Bluetooth Earbuds", "Electronics", "TechFlow", 5999, "earbuds",
		"Compact true wireless earbuds for everyday listening, workouts, and calls. IPX4 sweat resistance with 6-hour battery per charge. Seamless switching between phone, laptop, and tablet.",
		"versatile", nil},
	{"Noise-Cancelling Headphones", "Electronics", "TechFlow", 14999, "nc-headphones",
		"Professional-grade active noise cancellation for studio work and open-office environments. 30-hour battery, memory foam ear cushions for all-day comfort. Ideal for audio engineers and remote workers.",
		"intensive", nil},
	{"USB-C Charging Cable", "Electronics", "TechFlow", 1299, "cable",
		"Braided nylon USB-C cable rated for 100W power delivery and 10Gbps data transfer. 2-meter length with reinforced connectors. Works with laptops, tablets, phones, and peripherals.",
		"versatile", nil},
	{"USB-C Hub 7-in-1", "Electronics", "TechFlow", 4999, "usb-hub",
		"All-in-one docking solution with HDMI 4K, USB-A, USB-C PD, SD/microSD, and Ethernet. Aluminum body with thermal management for sustained workloads. Built for multi-monitor power users.",
		"intensive", nil},
	{"Wireless Charger Pad", "Electronics", "TechFlow", 2999, "charger-pad",
		"Slim Qi-compatible charging pad for nightstand or occasional desk use. 10W fast charge with LED indicator. No-fuss drop-and-charge convenience.",
		"occasional", nil},
	{"Fast Charging Adapter 65W", "Electronics", "TechFlow", 3499, "fast-charger",
		"GaN technology charger powering a laptop and phone simultaneously. Dual USB-C ports with intelligent power distribution. Compact enough for daily commute and travel.",
		"intensive", nil},
	{"Portable Power Bank 20000mAh", "Electronics", "TechFlow", 4499, "power-bank",
		"High-capacity battery pack with dual USB-C and USB-A output. Charges a laptop at 45W or three devices at once. Essential for travelers, field workers, and emergency backup.",
		"versatile", nil},
	{"Smart Watch", "Electronics", "NovaTech", 19999, "smart-watch",
		"Advanced health monitoring with ECG, SpO2, and continuous heart rate tracking. GPS, NFC payments, and 5-day battery. Designed for athletes and health-conscious professionals.",
		"intensive", nil},
	{"Fitness Tracker Band", "Electronics", "NovaTech", 6999, "fitness-tracker",
		"Simple step, sleep, and heart rate tracking in a lightweight band. 14-day battery life with smartphone notifications. Ideal for casual fitness goals and sleep awareness.",
		"occasional", nil},
	{"Wireless Mouse", "Electronics", "NovaTech", 2999, "wireless-mouse",
		"Ergonomic wireless mouse with adjustable DPI and silent clicks. Works on any surface with Bluetooth or USB receiver. Equally suited for office work, browsing, and light gaming.",
		"versatile", nil},
	{"Webcam 1080p", "Electronics", "NovaTech", 5999, "webcam",
		"Full HD webcam with auto-focus and built-in noise-reducing microphone. Easy clip-on mount for laptops and monitors. Great for occasional video calls and virtual meetings.",
		"occasional", nil},
	{"Studio Monitor Speakers", "Audio", "SoundWave", 24999, "studio-monitors",
		"Bi-amplified nearfield monitors with flat frequency response for mixing and mastering. XLR and TRS inputs with room correction controls. The reference standard for professional studios.",
		"intensive", []string{"US", "GB", "FR", "DE", "IT", "ES"}},
	{"Portable Bluetooth Speaker", "Audio", "SoundWave", 7999, "bt-speaker",
		"Waterproof IPX7 speaker for outdoor adventures and desktop use. 12-hour battery with 360-degree sound. Equally at home at a pool party or on a desk.",
		"versatile", nil},
	{"Soundbar 2.1", "Audio", "SoundWave", 14999, "soundbar",
		"Wireless subwoofer included for movie nights and casual TV watching. HDMI ARC, Bluetooth, and optical inputs. Easy setup without complex speaker placement.",
		"occasional", nil},
	{"Podcast Microphone USB", "Audio", "SoundWave", 8999, "podcast-mic",
		"Cardioid condenser with zero-latency monitoring and gain control. USB-C plug-and-play with shock mount included. Built for daily podcast recording and live streaming.",
		"intensive", []string{"US", "GB", "FR", "DE", "IT", "ES"}},
	{"Clip-On Lavalier Mic", "Audio", "SoundWave", 2999, "lav-mic",
		"Discreet clip-on microphone for interviews and presentations. 3.5mm jack compatible with cameras and smartphones. Reliable pickup without complex setup.",
		"occasional", nil},
	{"DAC/Amp Combo", "Audio", "SoundWave", 12999, "dac-amp",
		"High-resolution 32-bit/384kHz DAC with balanced headphone output. Drives demanding planar magnetic headphones effortlessly. Essential for critical listening and audiophile setups.",
		"intensive", []string{"US", "GB", "FR", "DE", "IT", "ES"}},
	{"In-Ear Monitors", "Audio", "AudioPrime", 9999, "iem",
		"Triple-driver IEMs with detachable cables and multiple ear tip sizes. 26dB noise isolation for stage and commute use. Balanced sound for musicians, audiophiles, and travelers.",
		"versatile", nil},
	{"Open-Back Headphones", "Audio", "AudioPrime", 17999, "open-back",
		"Audiophile open-back design with wide soundstage for extended listening sessions. Velour ear pads and lightweight frame reduce fatigue. Reference-grade for mixing and critical evaluation.",
		"intensive", nil},
	{"Wireless Turntable", "Audio", "AudioPrime", 19999, "turntable",
		"Belt-drive turntable with Bluetooth streaming and built-in preamp. Play vinyl records wirelessly through any Bluetooth speaker. Perfect for weekend listening and vinyl collectors.",
		"occasional", nil},
	{"Noise Machine", "Audio", "AudioPrime", 3999, "noise-machine",
		"30 natural soundscapes for sleep, focus, and relaxation. Compact design with timer and memory function. Useful in bedrooms, offices, and nurseries.",
		"versatile", nil},
	{"Phone Case (Universal)", "Accessories", "GearUp", 1999, "phone-case",
		"Shock-absorbent TPU case with raised edges for camera protection. Compatible with wireless charging and MagSafe. Everyday protection for any lifestyle.",
		"versatile", nil},
	{"Tablet Sleeve 11\"", "Accessories", "GearUp", 2499, "tablet-sleeve",
		"Padded neoprene sleeve with water-resistant zipper. Slim profile slides into bags and backpacks. Light protection for travel and storage.",
		"occasional", nil},
	{"Cable Organizer Kit", "Accessories", "GearUp", 1499, "cable-organizer",
		"Set of 10 reusable silicone ties and a travel pouch. Color-coded for quick identification of cables. Tames desk clutter and travel bags alike.",
		"versatile", nil},
	{"Laptop Backpack", "Accessories", "GearUp", 5999, "laptop-backpack",
		"30L capacity with padded laptop compartment, USB charging port, and luggage strap. Water-resistant fabric with ergonomic back panel. Designed for daily commuters and business travelers.",
		"intensive", nil},
	{"Screen Protector Pack", "Accessories", "GearUp", 999, "screen-protector",
		"Tempered glass 2-pack with alignment frame for bubble-free installation. 9H hardness and oleophobic coating. Replace as needed without worry.",
		"occasional", nil},
	{"Stylus Pen", "Accessories", "GearUp", 2999, "stylus",
		"4096 pressure levels with tilt sensitivity and palm rejection. Magnetic attachment and USB-C rechargeable. Built for digital artists and note-takers who draw daily.",
		"intensive", nil},
	{"Wireless Keyboard Case", "Accessories", "CarryTech", 4999, "keyboard-case",
		"Backlit keyboard with protective folio for tablets. Bluetooth 5.0 with 3-month battery between charges. Turns a tablet into a productivity workstation or typing companion.",
		"versatile", nil},
	{"AirTag Wallet", "Accessories", "CarryTech", 3499, "airtag-wallet",
		"Genuine leather bifold with hidden AirTag slot. RFID-blocking card pockets and slim profile. Keep track of your wallet without thinking about it.",
		"occasional", nil},
	{"Camera Strap", "Accessories", "CarryTech", 1999, "camera-strap",
		"Adjustable sling strap with quick-release buckle and anti-slip pad. Distributes weight for comfortable carrying. Great for day trips and casual photography.",
		"occasional", nil},
	{"Laptop Stand Adjustable", "Home Office", "DeskCraft", 4599, "laptop-stand",
		"Aluminum stand with 6 height settings for ergonomic screen positioning. Ventilated design prevents overheating during long work sessions. A must for daily laptop workstations.",
		"intensive", nil},
	{"Monitor Arm Dual", "Home Office", "DeskCraft", 8999, "monitor-arm",
		"Heavy-duty dual arm with full articulation and cable management. Supports monitors up to 32 inches and 20 lbs each. Frees desk space for intensive multi-monitor setups.",
		"intensive", nil},
	{"Ergonomic Wrist Rest", "Home Office", "DeskCraft", 1999, "wrist-rest",
		"Memory foam wrist rest with cooling gel layer. Non-slip base fits standard and tenkeyless keyboards. Reduces strain during both marathon coding sessions and casual browsing.",
		"versatile", nil},
	{"Desk Organizer Tray", "Home Office", "DeskCraft", 2499, "desk-organizer",
		"Bamboo desk tray with compartments for pens, cards, and phone. Minimalist design keeps essentials within reach. A tidy addition for light home office use.",
		"occasional", nil},
	{"Under-Desk Cable Tray", "Home Office", "DeskCraft", 2999, "cable-tray",
		"Steel mesh tray with clamp mount for hiding power strips and cables. No drilling required, fits desks up to 2 inches thick. Clean look for any workspace.",
		"versatile", nil},
	{"Sit-Stand Desk Converter", "Home Office", "DeskCraft", 29999, "sit-stand",
		"Spring-assisted riser with 35-inch surface for dual monitors and keyboard. Transitions smoothly between sitting and standing. Designed for all-day office use and health-conscious workers.",
		"intensive", nil},
	{"Desk Mat XL", "Home Office", "WorkZone", 3499, "desk-mat",
		"90x40cm microfiber desk mat with stitched edges and non-slip base. Protects desk surface and cushions wrists. Works for office, gaming, or crafting.",
		"versatile", nil},
	{"Footrest Ergonomic", "Home Office", "WorkZone", 4999, "footrest",
		"Tilting platform with textured surface for under-desk comfort. Encourages micro-movements to improve circulation. Helpful for those adapting to a new desk setup.",
		"occasional", nil},
	{"Monitor Light Bar", "Home Office", "WorkZone", 5999, "monitor-light",
		"Asymmetric LED bar illuminating the desk without screen glare. Auto-dimming sensor and adjustable color temperature. Essential for late-night work and eye strain reduction.",
		"intensive", nil},
	{"Whiteboard Desktop", "Home Office", "WorkZone", 2499, "whiteboard",
		"Tempered glass whiteboard with markers and eraser included. Frameless design sits flat on any desk. Handy for brainstorming sessions and quick to-do lists.",
		"occasional", nil},
	{"LED Desk Lamp", "Lighting", "LuxLight", 3499, "desk-lamp",
		"Touch-controlled lamp with 5 brightness levels and 3 color temperatures. Flexible gooseneck with USB charging port. Designed for extended reading and detail work.",
		"intensive", nil},
	{"RGB Light Strip 2m", "Lighting", "LuxLight", 1999, "light-strip",
		"Adhesive LED strip with remote control and 16 color modes. Easy peel-and-stick installation behind monitors or shelves. Fun accent lighting for movie nights and parties.",
		"occasional", nil},
	{"Smart Bulb E27 (2-pack)", "Lighting", "LuxLight", 2999, "smart-bulb",
		"WiFi-enabled bulbs with 16 million colors and voice control. Compatible with Alexa, Google Home, and HomeKit. Automate routines or set the mood for any occasion.",
		"versatile", nil},
	{"Ring Light 10\"", "Lighting", "LuxLight", 3999, "ring-light",
		"Adjustable tripod ring light with phone holder and 3 light modes. Flicker-free LEDs for video recording and live streams. Essential gear for content creators and professionals.",
		"intensive", nil},
	{"Ambient Light Panel", "Lighting", "LuxLight", 7999, "ambient-panel",
		"Modular hexagonal panels with touch activation and music sync. 16 million colors with wall-mount kit included. Eye-catching decor for gaming rooms and lounges.",
		"occasional", nil},
	{"Clip-On Book Light", "Lighting", "BrightLine", 999, "book-light",
		"Rechargeable LED light with warm and cool modes. Flexible neck clips to books, e-readers, and laptops. Handy for bedtime reading and travel.",
		"versatile", nil},
	{"Motion Sensor Night Light", "Lighting", "BrightLine", 1499, "night-light",
		"Warm white LED with PIR sensor and auto-off timer. Plugs directly into any outlet. Guides you safely at night in hallways, bathrooms, and bedrooms.",
		"versatile", nil},
	{"Sunset Lamp Projector", "Lighting", "BrightLine", 2499, "sunset-lamp",
		"Rotating projection lamp casting a warm sunset glow on walls. USB-powered with adjustable angle. Creates a relaxing ambiance for photos and winding down.",
		"occasional", nil},
	{"Mechanical Keyboard RGB", "Gaming", "GameEdge", 12999, "mech-keyboard",
		"Hot-swappable switches with per-key RGB and aluminum frame. N-key rollover with programmable macros. Built to endure daily gaming marathons and competitive play.",
		"intensive", nil},
	{"Gaming Mouse 16000 DPI", "Gaming", "GameEdge", 5999, "gaming-mouse",
		"Lightweight mouse with adjustable weight system and onboard memory. 1ms polling rate with optical sensor. Tuned for competitive FPS and intensive daily gaming.",
		"intensive", nil},
	{"Mouse Pad XXL", "Gaming", "GameEdge", 2499, "mousepad-xxl",
		"900x400mm micro-weave surface with anti-fray stitched edges. Fits full keyboard and mouse with room to spare. Works for gaming, office, and creative work.",
		"versatile", nil},
	{"Controller Stand", "Gaming", "GameEdge", 1999, "controller-stand",
		"Dual controller display stand with anti-scratch silicone cradles. Compact footprint for shelf or desk. Keeps controllers organized between gaming sessions.",
		"occasional", nil},
	{"Gaming Headset 7.1", "Gaming", "GameEdge", 8999, "gaming-headset",
		"Virtual 7.1 surround with retractable boom mic and inline controls. Breathable leatherette ear cups for long sessions. Engineered for competitive multiplayer and team communication.",
		"intensive", nil},
	{"Stream Deck Mini", "Gaming", "PixelForge", 7999, "stream-deck",
		"6 customizable LCD keys for OBS, Twitch, and productivity shortcuts. Drag-and-drop configuration with plugin ecosystem. A daily tool for streamers and power users.",
		"intensive", nil},
	{"Capture Card 4K", "Gaming", "PixelForge", 14999, "capture-card",
		"4K60 passthrough with 1080p60 capture via USB 3.0. Low-latency preview for console streaming. Plug-and-play solution for occasional streaming and recording.",
		"occasional", []string{"US", "CA", "GB", "JP"}},
	{"Racing Wheel", "Gaming", "PixelForge", 24999, "racing-wheel",
		"Force feedback wheel with 900-degree rotation and metal pedals. Desk clamp mount with quick-release hub. Enhances weekend racing sim sessions.",
		"occasional", []string{"US", "CA", "GB", "JP"}},
	{"VR Headset Stand", "Gaming", "PixelForge", 2999, "vr-stand",
		"Universal VR headset and controller display stand. Ventilated design keeps lenses clear. Protects and showcases your VR gear when not in use.",
		"versatile", nil},
	{"SSD External 1TB", "Storage", "DataVault", 8999, "ssd-1tb",
		"NVMe SSD with 1050MB/s read speeds and USB-C 3.2 Gen 2. Shock-resistant aluminum shell with encryption support. Reliable daily driver for video editors and developers.",
		"intensive", []string{"US", "CA", "GB", "DE", "FR"}},
	{"SSD External 2TB", "Storage", "DataVault", 14999, "ssd-2tb",
		"High-capacity NVMe drive for large media libraries and backups. Sequential reads up to 1050MB/s with hardware encryption. Essential for creative professionals managing terabytes daily.",
		"intensive", []string{"US", "CA", "GB", "DE", "FR"}},
	{"USB Flash Drive 128GB", "Storage", "DataVault", 1499, "flash-128",
		"Metal unibody design with retractable USB-A connector. USB 3.0 speeds up to 150MB/s. Convenient for file sharing and occasional transfers.",
		"occasional", nil},
	{"MicroSD Card 256GB", "Storage", "DataVault", 2999, "microsd-256",
		"A2-rated card with 160MB/s read speed and V30 video class. Includes SD adapter for cameras and laptops. Reliable storage for action cameras, drones, and phones.",
		"versatile", nil},
	{"NAS Enclosure 2-Bay", "Storage", "DataVault", 16999, "nas-2bay",
		"Dual-bay NAS with RAID 1 mirroring and remote access. ARM processor with 1GB RAM for home and small office use. Always-on network storage for backups and media serving.",
		"intensive", []string{"US", "CA", "GB", "DE", "FR"}},
	{"Hard Drive Dock", "Storage", "ByteKeep", 3999, "hdd-dock",
		"Dual-bay USB 3.0 dock supporting 2.5 and 3.5-inch SATA drives. Tool-free top-loading with offline clone function. Handy for data recovery, migration, and testing.",
		"versatile", nil},
	{"Memory Card Reader", "Storage", "ByteKeep", 1999, "card-reader",
		"USB-C reader supporting SD, microSD, CF, and MS formats. Compact aluminum body with status LED. Read cards as needed without adapters.",
		"occasional", nil},
	{"Encrypted USB Drive", "Storage", "ByteKeep", 4999, "encrypted-usb",
		"Hardware-encrypted USB drive with keypad PIN entry. FIPS 140-2 Level 3 certified with tamper-proof design. Protects sensitive data for business, personal, and field use.",
		"versatile", nil},
}

var featuredSlugs = map[string]bool{
	"headphones": true, "mech-keyboard": true, "laptop-stand": true,
	"bt-speaker": true, "ssd-1tb": true, "desk-lamp": true,
	"gaming-mouse": true, "smart-watch": true,
}

type catalogStore struct {
	Products   []icatalog.Product
	ProductSeq int
}

func newCatalogStore() *catalogStore {
	return &catalogStore{}
}

func (c *catalogStore) Find(id string) *icatalog.Product {
	for i := range c.Products {
		if c.Products[i].ID == id {
			return &c.Products[i]
		}
	}
	return nil
}

func (c *catalogStore) Filter(category, brand, query, usageType, country, currency, language string) []icatalog.Product {
	var result []icatalog.Product
	for _, p := range c.Products {
		if category != "" && !strings.EqualFold(p.Category, category) {
			continue
		}
		if brand != "" && !strings.EqualFold(p.Brand, brand) {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(p.Title), strings.ToLower(query)) {
			continue
		}
		if usageType != "" && !strings.EqualFold(p.UsageType, usageType) {
			continue
		}
		if country != "" && len(p.AvailableCountries) > 0 {
			if !icatalog.ContainsCountry(p.AvailableCountries, country) {
				continue
			}
		}
		result = append(result, p)
	}
	return result
}

func (c *catalogStore) CategoryCount() []icatalog.CategoryStat {
	counts := map[string]int{}
	order := []string{}
	for _, p := range c.Products {
		if _, seen := counts[p.Category]; !seen {
			order = append(order, p.Category)
		}
		counts[p.Category]++
	}
	result := make([]icatalog.CategoryStat, 0, len(order))
	for _, name := range order {
		result = append(result, icatalog.CategoryStat{
			Name:  name,
			Count: counts[name],
		})
	}
	return result
}

func (c *catalogStore) Lookup(id string, shipsTo string) *icatalog.Product {
	p := c.Find(id)
	if p == nil {
		return nil
	}
	if shipsTo != "" && len(p.AvailableCountries) > 0 {
		if !icatalog.ContainsCountry(p.AvailableCountries, shipsTo) {
			return nil
		}
	}
	return p
}

func (c *catalogStore) Search(params icatalog.SearchParams) []icatalog.SearchResult {
	limit := params.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 300 {
		limit = 300
	}

	query := strings.ToLower(params.Query)
	var results []icatalog.SearchResult
	for _, p := range c.Products {
		if query != "" {
			titleMatch := strings.Contains(strings.ToLower(p.Title), query)
			descMatch := strings.Contains(strings.ToLower(p.Description), query)
			catMatch := strings.Contains(strings.ToLower(p.Category), query)
			if !titleMatch && !descMatch && !catMatch {
				continue
			}
		}
		if params.MinPrice > 0 && p.Price < params.MinPrice {
			continue
		}
		if params.MaxPrice > 0 && p.Price > params.MaxPrice {
			continue
		}
		if params.AvailableForSale && p.Quantity <= 0 {
			continue
		}
		if params.ShipsTo != "" && len(p.AvailableCountries) > 0 {
			if !icatalog.ContainsCountry(p.AvailableCountries, params.ShipsTo) {
				continue
			}
		}
		results = append(results, icatalog.SearchResult{
			Product: p,
			InStock: p.Quantity > 0,
		})
		if len(results) >= limit {
			break
		}
	}
	return results
}

func (c *catalogStore) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var products []icatalog.Product
	if err := json.Unmarshal(data, &products); err != nil {
		return err
	}
	c.Products = products
	c.ProductSeq = len(products)
	return nil
}

func (c *catalogStore) Init(seed int64) {
	r := rand.New(rand.NewSource(seed))

	c.Products = make([]icatalog.Product, 0, len(templates))
	for i, t := range templates {
		factor := 0.80 + r.Float64()*0.40
		price := int(math.Round(float64(t.price) * factor))
		qty := 10 + r.Intn(191)

		rank := 100
		if featuredSlugs[t.slug] {
			rank = 10
		}

		c.Products = append(c.Products, icatalog.Product{
			ID:                 fmt.Sprintf("SKU-%03d", i+1),
			Title:              t.title,
			Category:           t.category,
			Brand:              t.brand,
			Price:              price,
			Quantity:           qty,
			Rank:               rank,
			ImageURL:           fmt.Sprintf("https://picsum.photos/seed/%s/200", t.slug),
			Description:        t.description,
			UsageType:          t.usageType,
			AvailableCountries: t.availableCountries,
		})
	}
	c.ProductSeq = len(c.Products)
}
