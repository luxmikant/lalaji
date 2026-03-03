"use client";

import { useState, FormEvent } from "react";

// â”€â”€ Seed data catalogue (mirrors DB migrations/000002_seed_data.up.sql) â”€â”€â”€â”€â”€â”€

const SELLERS = [
  { id: 1, name: "Nestle India Distributor", owner: "Rahul Mehta", city: "Bengaluru", state: "Karnataka", emoji: "ðŸ­", tag: "FMCG" },
  { id: 2, name: "Premium Rice Traders",     owner: "Suresh Patil",  city: "Mumbai",    state: "Maharashtra", emoji: "ðŸŒ¾", tag: "Staples" },
  { id: 3, name: "Gujarat Sugar Mills",      owner: "Bhavesh Patel", city: "Ahmedabad", state: "Gujarat",     emoji: "ðŸª", tag: "Commodities" },
];

const PRODUCTS = [
  { id: 1, sellerId: 1, name: "Maggi 500g Packet",    sku: "NES-MAG-500",  category: "Snacks",    mrp: 14,    price: 10,   weight: 0.5,  emoji: "ðŸœ", min: 24, fragile: false, perishable: false },
  { id: 2, sellerId: 1, name: "Nescafe Classic 200g", sku: "NES-COF-200",  category: "Beverages", mrp: 350,   price: 280,  weight: 0.25, emoji: "â˜•", min: 12, fragile: true,  perishable: false },
  { id: 3, sellerId: 2, name: "Basmati Rice 10Kg",    sku: "PRT-RIC-10K",  category: "Rice",      mrp: 700,   price: 500,  weight: 10,   emoji: "ðŸš", min: 5,  fragile: false, perishable: false },
  { id: 4, sellerId: 2, name: "Sona Masoori Rice 25Kg", sku: "PRT-RIC-25K", category: "Rice",    mrp: 1200,  price: 950,  weight: 25,   emoji: "ðŸŒ¾", min: 3,  fragile: false, perishable: false },
  { id: 5, sellerId: 3, name: "White Sugar 25Kg",     sku: "GSM-SUG-25K",  category: "Sugar",     mrp: 900,   price: 700,  weight: 25,   emoji: "ðŸ¬", min: 5,  fragile: false, perishable: false },
  { id: 6, sellerId: 3, name: "Jaggery Powder 5Kg",   sku: "GSM-JAG-5K",   category: "Sugar",     mrp: 400,   price: 320,  weight: 5,    emoji: "ðŸŸ«", min: 10, fragile: false, perishable: false },
];

const CUSTOMERS = [
  { id: 1, name: "Shree Kirana Store",   owner: "Ramesh Kumar",  city: "Bengaluru", state: "Karnataka",   pincode: "560034", emoji: "ðŸ¬", type: "Grocery" },
  { id: 2, name: "Andheri Mini Mart",    owner: "Sunil Sharma",  city: "Mumbai",    state: "Maharashtra", pincode: "400053", emoji: "ðŸ›’", type: "General" },
  { id: 3, name: "Dilli Grocery Hub",    owner: "Pankaj Verma",  city: "New Delhi", state: "Delhi",       pincode: "110001", emoji: "ðŸª", type: "Grocery" },
  { id: 4, name: "Hyderabad Fresh Mart", owner: "Lakshmi Reddy", city: "Hyderabad", state: "Telangana",   pincode: "500034", emoji: "ðŸŒ¿", type: "Dairy & Fresh" },
  { id: 5, name: "Chennai Bazaar",       owner: "Murugan S",     city: "Chennai",   state: "Tamil Nadu",  pincode: "600017", emoji: "ðŸŽª", type: "General" },
];

const TRANSPORT_META: Record<string, { label: string; emoji: string; color: string; dist: string }> = {
  minivan:   { label: "Mini Van",   emoji: "ðŸš", color: "text-blue-600 bg-blue-50",   dist: "0 â€“ 100 km" },
  truck:     { label: "Truck",      emoji: "ðŸš›", color: "text-purple-600 bg-purple-50", dist: "100 â€“ 500 km" },
  aeroplane: { label: "Aeroplane",  emoji: "âœˆï¸", color: "text-sky-600 bg-sky-50",      dist: "500+ km" },
};

// â”€â”€ Types â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

interface Breakdown {
  distanceKm: number;
  transportMode: string;
  ratePerKmPerKg: number;
  billableWeightKg: number;
  baseCourierCharge: number;
  distanceCharge: number;
  expressCharge: number;
  totalCharge: number;
}

interface NearestWarehouse {
  warehouseId: number;
  warehouseName: string;
  warehouseLocation: { lat: number; lng: number };
  distanceKm: number;
}

interface CalculationResult {
  shippingCharge: number;
  breakdown: Breakdown;
  nearestWarehouse: NearestWarehouse;
}

interface ApiResponse {
  success: boolean;
  data?: CalculationResult;
  error?: string;
  fields?: Record<string, string>;
}

// â”€â”€ Helpers â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";
const fmt = (n: number) =>
  n.toLocaleString("en-IN", { minimumFractionDigits: 2, maximumFractionDigits: 2 });

// â”€â”€ Main Component â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

export default function Home() {
  const [selectedSellerId, setSelectedSellerId] = useState<number | null>(null);
  const [selectedProductId, setSelectedProductId] = useState<number | null>(null);
  const [selectedCustomerId, setSelectedCustomerId] = useState<number | null>(null);
  const [deliverySpeed, setDeliverySpeed] = useState<"standard" | "express">("standard");

  const [result, setResult] = useState<CalculationResult | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const seller = SELLERS.find((s) => s.id === selectedSellerId) ?? null;
  const product = PRODUCTS.find((p) => p.id === selectedProductId) ?? null;
  const customer = CUSTOMERS.find((c) => c.id === selectedCustomerId) ?? null;
  const sellerProducts = PRODUCTS.filter((p) => p.sellerId === selectedSellerId);

  const canSubmit = selectedSellerId && selectedProductId && selectedCustomerId;

  const handleSellerSelect = (id: number) => {
    setSelectedSellerId(id);
    setSelectedProductId(null);
    setResult(null);
    setError(null);
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!canSubmit) return;
    setError(null);
    setResult(null);
    setLoading(true);

    try {
      const res = await fetch(`${API_URL}/api/v1/shipping-charge/calculate`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          sellerId: selectedSellerId,
          productId: selectedProductId,
          customerId: selectedCustomerId,
          deliverySpeed,
        }),
      });
      const json: ApiResponse = await res.json();
      if (!json.success || !json.data) {
        setError(json.error ?? "Unexpected error");
      } else {
        setResult(json.data);
      }
    } catch {
      setError("Could not reach the API. Check that the backend is running.");
    } finally {
      setLoading(false);
    }
  };

  const transportMeta = result ? (TRANSPORT_META[result.breakdown.transportMode] ?? null) : null;

  return (
    <div className="min-h-screen bg-[#FDF6EC]">
      {/* â”€â”€ Topbar â”€â”€ */}
      <header className="bg-white border-b border-orange-100 sticky top-0 z-10 shadow-sm">
        <div className="max-w-5xl mx-auto px-4 h-14 flex items-center justify-between">
          <div className="flex items-center gap-2.5">
            <span className="text-2xl">ðŸ“¦</span>
            <div>
              <span className="font-bold text-gray-900 text-base tracking-tight">Jambotails</span>
              <span className="text-gray-400 text-xs ml-2 hidden sm:inline">B2B Kirana Shipping Estimator</span>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <span className="text-xs text-gray-400 hidden sm:block">Powered by</span>
            <span className="text-xs bg-orange-100 text-orange-700 font-semibold px-2 py-0.5 rounded-full">Go + Next.js</span>
          </div>
        </div>
      </header>

      <main className="max-w-5xl mx-auto px-4 py-8 space-y-8">

        {/* â”€â”€ Hero â”€â”€ */}
        <div className="text-center space-y-2 pt-2">
          <h1 className="text-3xl font-extrabold text-gray-900 tracking-tight">
            Shipping Charge Estimator
          </h1>
          <p className="text-gray-500 max-w-xl mx-auto text-sm leading-relaxed">
            Select a seller, choose a product, and pick your Kirana store to get an accurate
            shipping estimate â€” including nearest warehouse routing and transport mode.
          </p>
        </div>

        {/* â”€â”€ Step 1: Seller â”€â”€ */}
        <Section step={1} title="Choose a Seller">
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            {SELLERS.map((s) => (
              <button
                key={s.id}
                onClick={() => handleSellerSelect(s.id)}
                className={`text-left rounded-2xl border-2 p-4 transition-all ${
                  selectedSellerId === s.id
                    ? "border-orange-500 bg-orange-50 shadow-md"
                    : "border-gray-200 bg-white hover:border-orange-300 hover:shadow-sm"
                }`}
              >
                <div className="text-3xl mb-2">{s.emoji}</div>
                <p className="font-semibold text-gray-800 text-sm leading-snug">{s.name}</p>
                <p className="text-xs text-gray-500 mt-0.5">{s.owner}</p>
                <div className="flex items-center gap-1 mt-2">
                  <span className="text-xs bg-gray-100 text-gray-600 px-2 py-0.5 rounded-full">{s.city}</span>
                  <span className="text-xs bg-orange-100 text-orange-700 px-2 py-0.5 rounded-full">{s.tag}</span>
                </div>
              </button>
            ))}
          </div>
        </Section>

        {/* â”€â”€ Step 2: Product â”€â”€ */}
        {selectedSellerId && (
          <Section step={2} title={`Products by ${seller?.name}`}>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
              {sellerProducts.map((p) => (
                <button
                  key={p.id}
                  onClick={() => { setSelectedProductId(p.id); setResult(null); setError(null); }}
                  className={`text-left rounded-2xl border-2 p-4 transition-all ${
                    selectedProductId === p.id
                      ? "border-orange-500 bg-orange-50 shadow-md"
                      : "border-gray-200 bg-white hover:border-orange-300 hover:shadow-sm"
                  }`}
                >
                  <div className="flex items-start justify-between mb-2">
                    <span className="text-3xl">{p.emoji}</span>
                    <span className="text-xs text-gray-400 bg-gray-100 px-2 py-0.5 rounded-full">{p.category}</span>
                  </div>
                  <p className="font-semibold text-gray-800 text-sm">{p.name}</p>
                  <p className="text-xs text-gray-400 mt-0.5 font-mono">{p.sku}</p>
                  <div className="flex items-center justify-between mt-3">
                    <div>
                      <span className="text-base font-bold text-gray-900">â‚¹{p.price}</span>
                      <span className="text-xs text-gray-400 line-through ml-1">â‚¹{p.mrp}</span>
                    </div>
                    <span className="text-xs text-gray-500">{p.weight} kg</span>
                  </div>
                  <div className="flex items-center gap-1 mt-2">
                    <span className="text-xs text-gray-500">Min. order: {p.min} units</span>
                    {p.fragile && <span className="text-xs bg-yellow-100 text-yellow-700 px-1.5 py-0.5 rounded-full">Fragile</span>}
                    {p.perishable && <span className="text-xs bg-red-100 text-red-700 px-1.5 py-0.5 rounded-full">Perishable</span>}
                  </div>
                </button>
              ))}
            </div>
          </Section>
        )}

        {/* â”€â”€ Step 3: Customer (Kirana Store) â”€â”€ */}
        {selectedProductId && (
          <Section step={3} title="Choose Kirana Store (Delivery To)">
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
              {CUSTOMERS.map((c) => (
                <button
                  key={c.id}
                  onClick={() => { setSelectedCustomerId(c.id); setResult(null); setError(null); }}
                  className={`text-left rounded-2xl border-2 p-4 transition-all ${
                    selectedCustomerId === c.id
                      ? "border-orange-500 bg-orange-50 shadow-md"
                      : "border-gray-200 bg-white hover:border-orange-300 hover:shadow-sm"
                  }`}
                >
                  <div className="text-3xl mb-2">{c.emoji}</div>
                  <p className="font-semibold text-gray-800 text-sm">{c.name}</p>
                  <p className="text-xs text-gray-500 mt-0.5">{c.owner}</p>
                  <div className="flex items-center gap-1 mt-2">
                    <span className="text-xs bg-gray-100 text-gray-600 px-2 py-0.5 rounded-full">{c.city}</span>
                    <span className="text-xs bg-green-100 text-green-700 px-2 py-0.5 rounded-full">{c.type}</span>
                  </div>
                  <p className="text-xs text-gray-400 mt-1">PIN: {c.pincode}</p>
                </button>
              ))}
            </div>
          </Section>
        )}

        {/* â”€â”€ Step 4: Delivery Speed + Submit â”€â”€ */}
        {selectedCustomerId && (
          <Section step={4} title="Delivery Speed">
            <form onSubmit={handleSubmit} className="space-y-5">
              <div className="flex flex-col sm:flex-row gap-3">
                {(["standard", "express"] as const).map((speed) => (
                  <button
                    key={speed}
                    type="button"
                    onClick={() => setDeliverySpeed(speed)}
                    className={`flex-1 rounded-2xl border-2 p-4 text-left transition-all ${
                      deliverySpeed === speed
                        ? "border-orange-500 bg-orange-50 shadow-md"
                        : "border-gray-200 bg-white hover:border-orange-300"
                    }`}
                  >
                    <div className="text-2xl mb-1">{speed === "standard" ? "ðŸ¢" : "âš¡"}</div>
                    <p className="font-semibold text-gray-800 capitalize">{speed}</p>
                    <p className="text-xs text-gray-500 mt-0.5">
                      {speed === "standard"
                        ? "â‚¹10 base + distance charge"
                        : "â‚¹10 base + distance charge + â‚¹1.2/kg surcharge"}
                    </p>
                  </button>
                ))}
              </div>

              {/* Summary strip */}
              <div className="bg-white rounded-2xl border border-gray-100 p-4 flex flex-wrap gap-3 text-sm text-gray-600">
                <Pill label="Seller" value={seller?.name ?? ""} emoji={seller?.emoji} />
                <span className="text-gray-200">|</span>
                <Pill label="Product" value={product?.name ?? ""} emoji={product?.emoji} />
                <span className="text-gray-200">|</span>
                <Pill label="Deliver to" value={customer?.name ?? ""} emoji={customer?.emoji} />
                <span className="text-gray-200">|</span>
                <Pill label="Speed" value={deliverySpeed} emoji={deliverySpeed === "express" ? "âš¡" : "ðŸ¢"} />
              </div>

              <button
                type="submit"
                disabled={loading}
                className="w-full bg-orange-500 hover:bg-orange-600 disabled:bg-orange-300 text-white font-bold px-8 py-3.5 rounded-2xl text-base transition shadow-lg shadow-orange-200 flex items-center justify-center gap-2"
              >
                {loading ? (
                  <>
                    <svg className="animate-spin h-5 w-5" viewBox="0 0 24 24" fill="none">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8z" />
                    </svg>
                    Calculating route &amp; chargesâ€¦
                  </>
                ) : (
                  "Get Shipping Estimate â†’"
                )}
              </button>
            </form>
          </Section>
        )}

        {/* â”€â”€ Error â”€â”€ */}
        {error && (
          <div className="bg-red-50 border border-red-200 rounded-2xl px-5 py-4 text-sm text-red-700 flex items-start gap-2">
            <span className="text-lg">âš ï¸</span>
            <div><strong>Error:</strong> {error}</div>
          </div>
        )}

        {/* â”€â”€ Results â”€â”€ */}
        {result && (
          <div className="space-y-5 animate-pulse-once">

            {/* Journey strip */}
            <div className="bg-white rounded-2xl border border-gray-100 shadow-sm p-5">
              <h3 className="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-4">Shipping Route</h3>
              <div className="flex items-center justify-between gap-2 overflow-x-auto">
                <RouteNode emoji={seller?.emoji ?? "ðŸ­"} label="Seller" sublabel={seller?.city ?? ""} />
                <RouteArrow label={`${fmt(result.nearestWarehouse.distanceKm)} km`} dim />
                <RouteNode emoji="ðŸ¢" label={result.nearestWarehouse.warehouseName} sublabel="Nearest Warehouse" highlight />
                <RouteArrow label={`${fmt(result.breakdown.distanceKm)} km`} />
                <RouteNode emoji={customer?.emoji ?? "ðŸ¬"} label={customer?.name ?? ""} sublabel={customer?.city ?? ""} />
              </div>
            </div>

            {/* Big charge card */}
            <div className="bg-gradient-to-r from-orange-500 to-amber-500 rounded-3xl p-6 text-white shadow-xl shadow-orange-200">
              <div className="flex items-start justify-between">
                <div>
                  <p className="text-sm font-medium text-orange-100">Total Shipping Charge</p>
                  <p className="text-5xl font-extrabold mt-0.5 tracking-tight">â‚¹{fmt(result.shippingCharge)}</p>
                </div>
                {transportMeta && (
                  <div className={`flex flex-col items-center rounded-xl px-4 py-3 bg-white/20`}>
                    <span className="text-3xl">{transportMeta.emoji}</span>
                    <span className="text-xs font-semibold mt-1 text-orange-50">{transportMeta.label}</span>
                    <span className="text-xs text-orange-200">{transportMeta.dist}</span>
                  </div>
                )}
              </div>
              <div className="flex flex-wrap gap-2 mt-4">
                <Badge text={deliverySpeed === "express" ? "âš¡ Express Delivery" : "ðŸ¢ Standard Delivery"} />
                <Badge text={`ðŸ“¦ ${product?.name}`} />
                <Badge text={`âš–ï¸ ${fmt(result.breakdown.billableWeightKg)} kg billable`} />
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
              {/* Breakdown */}
              <div className="bg-white rounded-2xl border border-gray-100 shadow-sm p-5">
                <h3 className="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-4">Charge Breakdown</h3>
                <table className="w-full text-sm">
                  <tbody>
                    <BreakdownRow label="Distance (WH â†’ Store)" value={`${fmt(result.breakdown.distanceKm)} km`} />
                    <BreakdownRow label="Billable Weight" value={`${fmt(result.breakdown.billableWeightKg)} kg`} note="max(actual, volumetric)" />
                    <BreakdownRow label="Rate" value={`â‚¹${result.breakdown.ratePerKmPerKg}/km/kg`} />
                    <BreakdownRow label="Base Courier Charge" value={`â‚¹${fmt(result.breakdown.baseCourierCharge)}`} />
                    <BreakdownRow label="Distance Charge" value={`â‚¹${fmt(result.breakdown.distanceCharge)}`} />
                    {result.breakdown.expressCharge > 0 && (
                      <BreakdownRow label="Express Surcharge" value={`â‚¹${fmt(result.breakdown.expressCharge)}`} accent />
                    )}
                    <tr className="border-t-2 border-gray-100">
                      <td className="pt-3 font-bold text-gray-900">Total</td>
                      <td className="pt-3 text-right font-extrabold text-orange-600 text-base tabular-nums">
                        â‚¹{fmt(result.breakdown.totalCharge)}
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>

              {/* Warehouse card */}
              <div className="bg-white rounded-2xl border border-gray-100 shadow-sm p-5 flex flex-col justify-between">
                <div>
                  <h3 className="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-4">Nearest Warehouse</h3>
                  <div className="flex items-center gap-3 mb-4">
                    <div className="w-12 h-12 rounded-xl bg-orange-100 flex items-center justify-center text-2xl">ðŸ¢</div>
                    <div>
                      <p className="font-bold text-gray-900">{result.nearestWarehouse.warehouseName}</p>
                      <p className="text-xs text-gray-500">Warehouse ID #{result.nearestWarehouse.warehouseId}</p>
                    </div>
                  </div>
                  <div className="grid grid-cols-2 gap-3 text-sm">
                    <InfoCell label="Seller â†’ WH" value={`${fmt(result.nearestWarehouse.distanceKm)} km`} />
                    <InfoCell label="WH â†’ Store" value={`${fmt(result.breakdown.distanceKm)} km`} />
                    <InfoCell label="Lat" value={result.nearestWarehouse.warehouseLocation.lat.toFixed(4) + "Â°N"} />
                    <InfoCell label="Lng" value={result.nearestWarehouse.warehouseLocation.lng.toFixed(4) + "Â°E"} />
                  </div>
                </div>
                <a
                  href={`https://www.google.com/maps?q=${result.nearestWarehouse.warehouseLocation.lat},${result.nearestWarehouse.warehouseLocation.lng}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="mt-4 flex items-center justify-center gap-2 border border-orange-200 text-orange-600 hover:bg-orange-50 rounded-xl px-4 py-2.5 text-sm font-semibold transition"
                >
                  ðŸ—ºï¸ View Warehouse on Maps â†—
                </a>
              </div>
            </div>

            {/* Formula note */}
            <p className="text-xs text-gray-400 text-center pb-4">
              Billable weight = max(actual_weight, LÃ—WÃ—HÃ·5000) Â· Total = â‚¹10 base + (rate Ã— dist Ã— weight)
              {result.breakdown.expressCharge > 0 ? " + â‚¹1.2 Ã— weight (express)" : ""}
            </p>
          </div>
        )}
      </main>
    </div>
  );
}

// â”€â”€ Sub-components â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

function Section({ step, title, children }: { step: number; title: string; children: React.ReactNode }) {
  return (
    <div className="space-y-3">
      <div className="flex items-center gap-2.5">
        <span className="w-7 h-7 rounded-full bg-orange-500 text-white text-xs font-bold flex items-center justify-center shrink-0">
          {step}
        </span>
        <h2 className="text-base font-bold text-gray-800">{title}</h2>
      </div>
      {children}
    </div>
  );
}

function Pill({ label, value, emoji }: { label: string; value: string; emoji?: string }) {
  return (
    <span className="flex items-center gap-1">
      {emoji && <span>{emoji}</span>}
      <span className="text-gray-400">{label}:</span>
      <span className="font-semibold text-gray-700 truncate max-w-[140px]">{value}</span>
    </span>
  );
}

function RouteNode({ emoji, label, sublabel, highlight, dim }: {
  emoji: string; label: string; sublabel: string; highlight?: boolean; dim?: boolean;
}) {
  return (
    <div className={`flex flex-col items-center gap-1 min-w-[80px] text-center ${dim ? "opacity-50" : ""}`}>
      <div className={`w-12 h-12 rounded-xl flex items-center justify-center text-2xl ${highlight ? "bg-orange-100 ring-2 ring-orange-400" : "bg-gray-100"}`}>
        {emoji}
      </div>
      <p className={`text-xs font-semibold ${highlight ? "text-orange-700" : "text-gray-700"} leading-tight max-w-[90px]`}>{label}</p>
      <p className="text-[10px] text-gray-400 leading-tight max-w-[90px]">{sublabel}</p>
    </div>
  );
}

function RouteArrow({ label, dim }: { label: string; dim?: boolean }) {
  return (
    <div className={`flex flex-col items-center gap-0.5 flex-1 min-w-[50px] ${dim ? "opacity-40" : ""}`}>
      <div className="w-full border-t-2 border-dashed border-orange-300 relative">
        <span className="absolute right-0 top-[-7px] text-orange-400">â–¶</span>
      </div>
      <span className="text-[10px] text-gray-400 font-mono">{label}</span>
    </div>
  );
}

function Badge({ text }: { text: string }) {
  return (
    <span className="text-xs bg-white/20 text-white px-2.5 py-1 rounded-full font-medium">
      {text}
    </span>
  );
}

function BreakdownRow({ label, value, note, accent, bold }: {
  label: string; value: string; note?: string; accent?: boolean; bold?: boolean;
}) {
  return (
    <tr className="border-b border-gray-50">
      <td className={`py-2 ${bold ? "font-bold text-gray-900" : "text-gray-500"}`}>
        {label}
        {note && <span className="block text-[10px] text-gray-400">{note}</span>}
      </td>
      <td className={`py-2 text-right tabular-nums font-medium ${accent ? "text-orange-600" : bold ? "text-gray-900 font-bold" : "text-gray-800"}`}>
        {value}
      </td>
    </tr>
  );
}

function InfoCell({ label, value }: { label: string; value: string }) {
  return (
    <div className="bg-gray-50 rounded-xl p-2.5">
      <p className="text-[10px] text-gray-400 uppercase tracking-wide">{label}</p>
      <p className="text-sm font-semibold text-gray-800 tabular-nums mt-0.5">{value}</p>
    </div>
  );
}

