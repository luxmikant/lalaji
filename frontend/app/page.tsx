"use client";

import { useState, FormEvent } from "react";

// -- Seed data catalogue (mirrors DB migrations/000002_seed_data.up.sql) ------

const SELLERS = [
  { id: 1, name: "Nestle India Distributor",  owner: "Rahul Mehta",   city: "Bengaluru",  state: "Karnataka",   emoji: "\u{1F3ED}", tag: "FMCG",        tagColor: "bg-violet-100 text-violet-700" },
  { id: 2, name: "Premium Rice Traders",      owner: "Suresh Patil",  city: "Mumbai",     state: "Maharashtra", emoji: "\u{1F33E}", tag: "Staples",     tagColor: "bg-emerald-100 text-emerald-700" },
  { id: 3, name: "Gujarat Sugar Mills",       owner: "Bhavesh Patel", city: "Ahmedabad",  state: "Gujarat",     emoji: "\u{1F3EA}", tag: "Commodities", tagColor: "bg-amber-100 text-amber-700" },
];

const PRODUCTS = [
  { id: 1, sellerId: 1, name: "Maggi 500g Packet",      sku: "NES-MAG-500",  category: "Snacks",    mrp: 14,   price: 10,  weight: 0.5,  emoji: "\u{1F35C}", min: 24, fragile: false, perishable: false },
  { id: 2, sellerId: 1, name: "Nescafe Classic 200g",   sku: "NES-COF-200",  category: "Beverages", mrp: 350,  price: 280, weight: 0.25, emoji: "\u{2615}",  min: 12, fragile: true,  perishable: false },
  { id: 3, sellerId: 2, name: "Basmati Rice 10Kg",      sku: "PRT-RIC-10K",  category: "Rice",      mrp: 700,  price: 500, weight: 10,   emoji: "\u{1F35A}", min: 5,  fragile: false, perishable: false },
  { id: 4, sellerId: 2, name: "Sona Masoori Rice 25Kg", sku: "PRT-RIC-25K",  category: "Rice",      mrp: 1200, price: 950, weight: 25,   emoji: "\u{1F33E}", min: 3,  fragile: false, perishable: false },
  { id: 5, sellerId: 3, name: "White Sugar 25Kg",       sku: "GSM-SUG-25K",  category: "Sugar",     mrp: 900,  price: 700, weight: 25,   emoji: "\u{1F36C}", min: 5,  fragile: false, perishable: false },
  { id: 6, sellerId: 3, name: "Jaggery Powder 5Kg",     sku: "GSM-JAG-5K",   category: "Sugar",     mrp: 400,  price: 320, weight: 5,    emoji: "\u{1F7EB}", min: 10, fragile: false, perishable: false },
];

const CUSTOMERS = [
  { id: 1, name: "Shree Kirana Store",   owner: "Ramesh Kumar",  city: "Bengaluru", state: "Karnataka",   pincode: "560034", emoji: "\u{1F3EC}", type: "Grocery" },
  { id: 2, name: "Andheri Mini Mart",    owner: "Sunil Sharma",  city: "Mumbai",    state: "Maharashtra", pincode: "400053", emoji: "\u{1F6D2}", type: "General" },
  { id: 3, name: "Dilli Grocery Hub",    owner: "Pankaj Verma",  city: "New Delhi", state: "Delhi",       pincode: "110001", emoji: "\u{1F3EA}", type: "Grocery" },
  { id: 4, name: "Hyderabad Fresh Mart", owner: "Lakshmi Reddy", city: "Hyderabad", state: "Telangana",   pincode: "500034", emoji: "\u{1F33F}", type: "Dairy & Fresh" },
  { id: 5, name: "Chennai Bazaar",       owner: "Murugan S",     city: "Chennai",   state: "Tamil Nadu",  pincode: "600017", emoji: "\u{1F3AA}", type: "General" },
];

const TRANSPORT_META: Record<string, { label: string; emoji: string; gradient: string; dist: string }> = {
  minivan:   { label: "Mini Van",  emoji: "\u{1F690}", gradient: "from-cyan-500 to-blue-600",     dist: "0 \u2013 100 km"  },
  truck:     { label: "Truck",     emoji: "\u{1F69B}", gradient: "from-violet-500 to-purple-700", dist: "100 \u2013 500 km" },
  aeroplane: { label: "Aeroplane", emoji: "\u2708\uFE0F", gradient: "from-sky-400 to-indigo-600", dist: "500+ km"          },
};

// -- Types --------------------------------------------------------------------

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

// -- Helpers ------------------------------------------------------------------

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";
const fmt = (n: number) =>
  n.toLocaleString("en-IN", { minimumFractionDigits: 2, maximumFractionDigits: 2 });

// -- Main Component -----------------------------------------------------------

export default function Home() {
  const [selectedSellerId,   setSelectedSellerId]   = useState<number | null>(null);
  const [selectedProductId,  setSelectedProductId]  = useState<number | null>(null);
  const [selectedCustomerId, setSelectedCustomerId] = useState<number | null>(null);
  const [deliverySpeed, setDeliverySpeed] = useState<"standard" | "express">("standard");

  const [result,  setResult]  = useState<CalculationResult | null>(null);
  const [error,   setError]   = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const seller   = SELLERS.find((s) => s.id === selectedSellerId)   ?? null;
  const product  = PRODUCTS.find((p) => p.id === selectedProductId) ?? null;
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

  const tMeta = result ? (TRANSPORT_META[result.breakdown.transportMode] ?? null) : null;

  return (
    <div className="min-h-screen bg-slate-50 font-sans">

      {/* Animated Dark Hero */}
      <div className="gradient-hero text-white">
        <div className="max-w-5xl mx-auto px-4 pt-10 pb-14">

          {/* Nav */}
          <div className="flex items-center justify-between mb-10">
            <div className="flex items-center gap-3">
              <div className="w-9 h-9 rounded-xl bg-white/10 backdrop-blur flex items-center justify-center text-xl animate-float">
                📦
              </div>
              <div>
                <span className="font-display font-bold text-lg tracking-wide">Lalaji</span>
                <span className="text-white/50 text-xs ml-2 hidden sm:inline">B2B Kirana Marketplace</span>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <a href={`${API_URL}/docs`} target="_blank"
                className="text-xs px-3 py-1.5 rounded-lg bg-white/10 hover:bg-white/20 transition text-white/80 font-medium border border-white/10">
                ⚡ API Docs
              </a>
              <span className="text-xs px-3 py-1.5 rounded-lg bg-violet-500/30 text-violet-200 font-semibold border border-violet-400/20">
                Go + Next.js
              </span>
            </div>
          </div>

          {/* Hero copy */}
          <div className="text-center space-y-4 animate-fade-in">
            <div className="inline-flex items-center gap-2 bg-white/10 backdrop-blur border border-white/15 rounded-full px-4 py-1.5 text-sm text-white/80 mb-2">
              <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse inline-block" />
              Seedha Source Se — Straight from Source
            </div>
            <h1 className="font-display font-bold text-5xl sm:text-6xl tracking-tight leading-none">
              Shipping Charge
              <span className="block bg-gradient-to-r from-violet-300 via-pink-300 to-amber-300 bg-clip-text text-transparent">
                Estimator
              </span>
            </h1>
            <p className="text-white/60 max-w-lg mx-auto text-base leading-relaxed">
              Find the nearest warehouse, pick your delivery speed, and get a complete
              shipping cost breakdown — in seconds.
            </p>
            <div className="flex flex-wrap justify-center gap-2 pt-2">
              {Object.values(TRANSPORT_META).map((m) => (
                <span key={m.label}
                  className="flex items-center gap-1.5 text-xs bg-white/10 border border-white/15 rounded-full px-3 py-1 text-white/70">
                  <span>{m.emoji}</span>{m.label}
                  <span className="text-white/40">·</span>
                  <span>{m.dist}</span>
                </span>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* Steps */}
      <main className="max-w-5xl mx-auto px-4 -mt-6 pb-16 space-y-5">

        {/* Step 1 - Seller */}
        <Step step={1} title="Choose a Seller" color="violet">
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            {SELLERS.map((s, i) => (
              <button
                key={s.id}
                onClick={() => handleSellerSelect(s.id)}
                className={[
                  "animate-fade-up text-left rounded-2xl border-2 p-4 transition-all duration-200",
                  `stagger-${i + 1}`,
                  selectedSellerId === s.id
                    ? "border-violet-500 bg-violet-50 shadow-lg shadow-violet-100 glow-selected"
                    : "border-slate-200 bg-white hover:border-violet-300 hover:shadow-md hover:-translate-y-0.5",
                ].join(" ")}
              >
                <div className="text-3xl mb-2">{s.emoji}</div>
                <p className={`font-semibold text-sm leading-snug ${selectedSellerId === s.id ? "text-violet-900" : "text-slate-800"}`}>
                  {s.name}
                </p>
                <p className="text-xs text-slate-500 mt-0.5">{s.owner}</p>
                <div className="flex items-center gap-1.5 mt-2 flex-wrap">
                  <span className="text-xs bg-slate-100 text-slate-600 px-2 py-0.5 rounded-full">{s.city}</span>
                  <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${s.tagColor}`}>{s.tag}</span>
                </div>
              </button>
            ))}
          </div>
        </Step>

        {/* Step 2 - Product */}
        {selectedSellerId && (
          <Step step={2} title={`Products by ${seller?.name}`} color="teal">
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
              {sellerProducts.map((p, i) => (
                <button
                  key={p.id}
                  onClick={() => { setSelectedProductId(p.id); setResult(null); setError(null); }}
                  className={[
                    "animate-fade-up text-left rounded-2xl border-2 p-4 transition-all duration-200",
                    `stagger-${i + 1}`,
                    selectedProductId === p.id
                      ? "border-teal-500 bg-teal-50 shadow-lg shadow-teal-100 glow-selected"
                      : "border-slate-200 bg-white hover:border-teal-300 hover:shadow-md hover:-translate-y-0.5",
                  ].join(" ")}
                >
                  <div className="flex items-start justify-between mb-2">
                    <span className="text-3xl">{p.emoji}</span>
                    <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${
                      selectedProductId === p.id ? "bg-teal-100 text-teal-700" : "bg-slate-100 text-slate-600"
                    }`}>{p.category}</span>
                  </div>
                  <p className={`font-semibold text-sm ${selectedProductId === p.id ? "text-teal-900" : "text-slate-800"}`}>
                    {p.name}
                  </p>
                  <p className="text-xs text-slate-400 mt-0.5 font-mono">{p.sku}</p>
                  <div className="flex items-center justify-between mt-3">
                    <div>
                      <span className="text-base font-bold text-slate-900">&#8377;{p.price}</span>
                      <span className="text-xs text-slate-400 line-through ml-1">&#8377;{p.mrp}</span>
                    </div>
                    <span className="text-xs bg-slate-50 text-slate-500 px-2 py-0.5 rounded-full border border-slate-100">{p.weight} kg</span>
                  </div>
                  <div className="flex items-center gap-1 mt-2 flex-wrap">
                    <span className="text-xs text-slate-400">Min {p.min} units</span>
                    {p.fragile    && <span className="text-xs bg-amber-100 text-amber-700 px-1.5 py-0.5 rounded-full font-medium">Fragile</span>}
                    {p.perishable && <span className="text-xs bg-rose-100 text-rose-700 px-1.5 py-0.5 rounded-full font-medium">Perishable</span>}
                  </div>
                </button>
              ))}
            </div>
          </Step>
        )}

        {/* Step 3 - Customer */}
        {selectedProductId && (
          <Step step={3} title="Choose Kirana Store &mdash; Delivery To" color="emerald">
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
              {CUSTOMERS.map((c, i) => (
                <button
                  key={c.id}
                  onClick={() => { setSelectedCustomerId(c.id); setResult(null); setError(null); }}
                  className={[
                    "animate-fade-up text-left rounded-2xl border-2 p-4 transition-all duration-200",
                    `stagger-${i % 5 + 1}`,
                    selectedCustomerId === c.id
                      ? "border-emerald-500 bg-emerald-50 shadow-lg shadow-emerald-100 glow-selected"
                      : "border-slate-200 bg-white hover:border-emerald-300 hover:shadow-md hover:-translate-y-0.5",
                  ].join(" ")}
                >
                  <div className="text-3xl mb-2">{c.emoji}</div>
                  <p className={`font-semibold text-sm ${selectedCustomerId === c.id ? "text-emerald-900" : "text-slate-800"}`}>
                    {c.name}
                  </p>
                  <p className="text-xs text-slate-500 mt-0.5">{c.owner}</p>
                  <div className="flex items-center gap-1.5 mt-2 flex-wrap">
                    <span className="text-xs bg-slate-100 text-slate-600 px-2 py-0.5 rounded-full">{c.city}</span>
                    <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${
                      selectedCustomerId === c.id ? "bg-emerald-100 text-emerald-700" : "bg-slate-100 text-slate-600"
                    }`}>{c.type}</span>
                  </div>
                  <p className="text-xs text-slate-400 mt-1 font-mono">PIN {c.pincode}</p>
                </button>
              ))}
            </div>
          </Step>
        )}

        {/* Step 4 - Speed + Submit */}
        {selectedCustomerId && (
          <Step step={4} title="Delivery Speed" color="amber">
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="flex flex-col sm:flex-row gap-3">
                {(["standard", "express"] as const).map((speed) => (
                  <button
                    key={speed}
                    type="button"
                    onClick={() => setDeliverySpeed(speed)}
                    className={[
                      "flex-1 rounded-2xl border-2 p-4 text-left transition-all duration-200",
                      deliverySpeed === speed
                        ? "border-amber-500 bg-amber-50 shadow-md"
                        : "border-slate-200 bg-white hover:border-amber-300 hover:shadow-sm",
                    ].join(" ")}
                  >
                    <div className="text-2xl mb-1.5">{speed === "standard" ? "🐢" : "⚡"}</div>
                    <p className={`font-semibold capitalize ${deliverySpeed === speed ? "text-amber-900" : "text-slate-800"}`}>
                      {speed}
                    </p>
                    <p className="text-xs text-slate-500 mt-0.5">
                      {speed === "standard"
                        ? "&#8377;10 base + distance charge"
                        : "&#8377;10 base + distance charge + &#8377;1.2 / kg express surcharge"}
                    </p>
                  </button>
                ))}
              </div>

              {/* Summary strip */}
              <div className="bg-white rounded-2xl border border-slate-100 p-4 flex flex-wrap gap-3 text-sm shadow-sm">
                <SummaryPill label="Seller"     value={seller?.name ?? ""}   emoji={seller?.emoji} />
                <Divider />
                <SummaryPill label="Product"    value={product?.name ?? ""}  emoji={product?.emoji} />
                <Divider />
                <SummaryPill label="Deliver to" value={customer?.name ?? ""} emoji={customer?.emoji} />
                <Divider />
                <SummaryPill label="Speed"      value={deliverySpeed}        emoji={deliverySpeed === "express" ? "⚡" : "🐢"} />
              </div>

              <button
                type="submit"
                disabled={loading}
                className={[
                  "w-full font-bold px-8 py-4 rounded-2xl text-base transition-all duration-200 shadow-lg flex items-center justify-center gap-2",
                  loading
                    ? "bg-violet-400 text-white cursor-not-allowed"
                    : "bg-gradient-to-r from-violet-600 to-indigo-700 hover:from-violet-700 hover:to-indigo-800 text-white shadow-violet-200 hover:-translate-y-0.5 active:translate-y-0",
                ].join(" ")}
              >
                {loading ? (
                  <>
                    <svg className="animate-spin h-5 w-5" viewBox="0 0 24 24" fill="none">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8z" />
                    </svg>
                    Calculating route &amp; charges…
                  </>
                ) : "Get Shipping Estimate →"}
              </button>
            </form>
          </Step>
        )}

        {/* Error */}
        {error && (
          <div className="animate-fade-up bg-rose-50 border border-rose-200 rounded-2xl px-5 py-4 text-sm text-rose-700 flex items-start gap-3 shadow-sm">
            <span className="text-xl shrink-0">⚠️</span>
            <div><strong className="font-semibold">Error: </strong>{error}</div>
          </div>
        )}

        {/* Results */}
        {result && (
          <div className="space-y-4 animate-fade-up">

            {/* Route strip */}
            <div className="bg-white rounded-2xl border border-slate-100 shadow-sm p-5">
              <h3 className="text-xs font-semibold text-slate-400 uppercase tracking-widest mb-4">Shipping Route</h3>
              <div className="flex items-center justify-between gap-2 overflow-x-auto pb-1">
                <RouteNode emoji={seller?.emoji ?? "🏭"} label="Seller"         sublabel={seller?.city ?? ""}                color="violet" />
                <RouteArrow label={`${fmt(result.nearestWarehouse.distanceKm)} km`} dim />
                <RouteNode emoji="🏢"                    label={result.nearestWarehouse.warehouseName} sublabel="Nearest WH" color="indigo" highlight />
                <RouteArrow label={`${fmt(result.breakdown.distanceKm)} km`} />
                <RouteNode emoji={customer?.emoji ?? "🏬"} label={customer?.name ?? ""} sublabel={customer?.city ?? ""}      color="emerald" />
              </div>
            </div>

            {/* Total charge card */}
            <div className={`rounded-3xl p-6 text-white shadow-2xl bg-gradient-to-br ${
              tMeta ? tMeta.gradient : "from-violet-600 to-indigo-800"
            }`}>
              <div className="flex items-start justify-between flex-wrap gap-4">
                <div className="animate-count-up">
                  <p className="text-sm font-medium text-white/70 mb-1">Total Shipping Charge</p>
                  <p className="font-display font-bold text-6xl tracking-tight leading-none">
                    &#8377;{fmt(result.shippingCharge)}
                  </p>
                  {tMeta && (
                    <p className="text-sm text-white/70 mt-2 flex items-center gap-1.5">
                      {tMeta.emoji} {tMeta.label} · {tMeta.dist}
                    </p>
                  )}
                </div>
                <div className="flex flex-wrap gap-2">
                  <ResultBadge text={deliverySpeed === "express" ? "⚡ Express" : "🐢 Standard"} />
                  <ResultBadge text={`📦 ${product?.name}`} />
                  <ResultBadge text={`⚖️ ${fmt(result.breakdown.billableWeightKg)} kg`} />
                </div>
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">

              {/* Breakdown */}
              <div className="bg-white rounded-2xl border border-slate-100 shadow-sm p-5">
                <h3 className="text-xs font-semibold text-slate-400 uppercase tracking-widest mb-4">Charge Breakdown</h3>
                <table className="w-full text-sm">
                  <tbody>
                    <BRow label="Distance (WH → Store)" value={`${fmt(result.breakdown.distanceKm)} km`} />
                    <BRow label="Billable Weight"            value={`${fmt(result.breakdown.billableWeightKg)} kg`} note="max(actual, volumetric)" />
                    <BRow label="Rate"                       value={`₹${result.breakdown.ratePerKmPerKg} / km / kg`} />
                    <BRow label="Base Courier Charge"        value={`₹${fmt(result.breakdown.baseCourierCharge)}`} />
                    <BRow label="Distance Charge"            value={`₹${fmt(result.breakdown.distanceCharge)}`} />
                    {result.breakdown.expressCharge > 0 && (
                      <BRow label="Express Surcharge (₹1.2/kg)" value={`₹${fmt(result.breakdown.expressCharge)}`} accent />
                    )}
                    <tr className="border-t-2 border-slate-100">
                      <td className="pt-3 font-bold text-slate-900">Total</td>
                      <td className="pt-3 text-right font-extrabold text-violet-700 text-base tabular-nums font-display">
                        &#8377;{fmt(result.breakdown.totalCharge)}
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>

              {/* Warehouse card */}
              <div className="bg-white rounded-2xl border border-slate-100 shadow-sm p-5 flex flex-col justify-between">
                <div>
                  <h3 className="text-xs font-semibold text-slate-400 uppercase tracking-widest mb-4">Nearest Warehouse</h3>
                  <div className="flex items-center gap-3 mb-4">
                    <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-violet-100 to-indigo-100 flex items-center justify-center text-2xl shadow-inner">
                      🏢
                    </div>
                    <div>
                      <p className="font-bold text-slate-900">{result.nearestWarehouse.warehouseName}</p>
                      <p className="text-xs text-slate-400">ID <span className="font-mono">#{result.nearestWarehouse.warehouseId}</span></p>
                    </div>
                  </div>
                  <div className="grid grid-cols-2 gap-2.5">
                    <InfoCell label="Seller → WH" value={`${fmt(result.nearestWarehouse.distanceKm)} km`} />
                    <InfoCell label="WH → Store"  value={`${fmt(result.breakdown.distanceKm)} km`} />
                    <InfoCell label="Latitude"         value={`${result.nearestWarehouse.warehouseLocation.lat.toFixed(4)}° N`} />
                    <InfoCell label="Longitude"        value={`${result.nearestWarehouse.warehouseLocation.lng.toFixed(4)}° E`} />
                  </div>
                </div>
                <a
                  href={`https://www.google.com/maps?q=${result.nearestWarehouse.warehouseLocation.lat},${result.nearestWarehouse.warehouseLocation.lng}`}
                  target="_blank" rel="noopener noreferrer"
                  className="mt-4 flex items-center justify-center gap-2 border-2 border-violet-200 text-violet-700 hover:bg-violet-50 rounded-xl px-4 py-2.5 text-sm font-semibold transition">
                  🗺️ View Warehouse on Maps ↗
                </a>
              </div>
            </div>

            <p className="text-xs text-slate-400 text-center pb-4 font-mono">
              Billable = max(actual, L×W×H÷5000) · Total = ₹10 + rate × dist × weight
              {result.breakdown.expressCharge > 0 ? " + ₹1.2 × weight (express)" : ""}
            </p>
          </div>
        )}
      </main>

      {/* Footer */}
      <footer className="border-t border-slate-200 bg-white py-6">
        <div className="max-w-5xl mx-auto px-4 flex flex-wrap items-center justify-between gap-3 text-xs text-slate-400">
          <span>© 2026 Lalaji — B2B Kirana Marketplace · Shipping Estimator v1.0</span>
          <div className="flex items-center gap-3">
            <a href={`${API_URL}/docs`} target="_blank" className="hover:text-violet-600 transition">API Docs ↗</a>
            <a href={`${API_URL}/api/openapi.yaml`} target="_blank" className="hover:text-violet-600 transition font-mono">openapi.yaml ↗</a>
          </div>
        </div>
      </footer>
    </div>
  );
}

// -- Sub-components -----------------------------------------------------------

const STEP_COLORS: Record<string, { ring: string; bg: string; text: string; number: string }> = {
  violet:  { ring: "ring-violet-200",  bg: "bg-violet-50",  text: "text-violet-700",  number: "bg-gradient-to-br from-violet-500 to-indigo-600"  },
  teal:    { ring: "ring-teal-200",    bg: "bg-teal-50",    text: "text-teal-700",    number: "bg-gradient-to-br from-teal-500 to-cyan-600"       },
  emerald: { ring: "ring-emerald-200", bg: "bg-emerald-50", text: "text-emerald-700", number: "bg-gradient-to-br from-emerald-500 to-teal-600"    },
  amber:   { ring: "ring-amber-200",   bg: "bg-amber-50",   text: "text-amber-700",   number: "bg-gradient-to-br from-amber-500 to-orange-500"    },
};

function Step({ step, title, color, children }: {
  step: number; title: string; color: string; children: React.ReactNode;
}) {
  const c = STEP_COLORS[color] ?? STEP_COLORS.violet;
  return (
    <div className={`rounded-3xl border ${c.ring} ${c.bg} p-5 shadow-sm`}>
      <div className="flex items-center gap-3 mb-4">
        <span className={`w-8 h-8 rounded-xl ${c.number} text-white text-sm font-bold flex items-center justify-center shadow-md shrink-0`}>
          {step}
        </span>
        <h2 className={`text-sm font-bold ${c.text} tracking-wide`}>{title}</h2>
      </div>
      {children}
    </div>
  );
}

function SummaryPill({ label, value, emoji }: { label: string; value: string; emoji?: string }) {
  return (
    <span className="flex items-center gap-1 min-w-0">
      {emoji && <span className="shrink-0">{emoji}</span>}
      <span className="text-slate-400 text-xs">{label}:</span>
      <span className="font-semibold text-slate-700 text-xs truncate max-w-[120px]">{value}</span>
    </span>
  );
}

function Divider() {
  return <span className="text-slate-200 text-xs hidden sm:inline">|</span>;
}

function RouteNode({ emoji, label, sublabel, color, highlight, dim }: {
  emoji: string; label: string; sublabel: string;
  color: string; highlight?: boolean; dim?: boolean;
}) {
  const colorMap: Record<string, string> = {
    violet:  "bg-violet-100 ring-violet-400",
    indigo:  "bg-indigo-100 ring-indigo-500",
    emerald: "bg-emerald-100 ring-emerald-400",
  };
  return (
    <div className={`flex flex-col items-center gap-1 min-w-[76px] text-center ${dim ? "opacity-50" : ""}`}>
      <div className={`w-12 h-12 rounded-xl flex items-center justify-center text-2xl ${colorMap[color] ?? "bg-slate-100"} ${highlight ? "ring-2 shadow-md" : ""}`}>
        {emoji}
      </div>
      <p className="text-xs font-semibold text-slate-700 leading-tight max-w-[88px]">{label}</p>
      <p className="text-[10px] text-slate-400 leading-tight max-w-[88px]">{sublabel}</p>
    </div>
  );
}

function RouteArrow({ label, dim }: { label: string; dim?: boolean }) {
  return (
    <div className={`flex flex-col items-center gap-0.5 flex-1 min-w-[44px] ${dim ? "opacity-40" : ""}`}>
      <div className="w-full border-t-2 border-dashed border-violet-300 relative">
        <span className="absolute right-0 top-[-8px] text-violet-400 text-sm">&#9654;</span>
      </div>
      <span className="text-[10px] text-slate-400 font-mono">{label}</span>
    </div>
  );
}

function ResultBadge({ text }: { text: string }) {
  return (
    <span className="text-xs bg-white/20 backdrop-blur text-white px-3 py-1 rounded-full font-medium border border-white/20">
      {text}
    </span>
  );
}

function BRow({ label, value, note, accent }: {
  label: string; value: string; note?: string; accent?: boolean;
}) {
  return (
    <tr className="border-b border-slate-50">
      <td className="py-2 text-slate-500 text-xs">
        {label}
        {note && <span className="block text-[10px] text-slate-400 italic">{note}</span>}
      </td>
      <td className={`py-2 text-right tabular-nums font-medium text-xs ${accent ? "text-violet-600 font-semibold" : "text-slate-800"}`}>
        {value}
      </td>
    </tr>
  );
}

function InfoCell({ label, value }: { label: string; value: string }) {
  return (
    <div className="bg-slate-50 rounded-xl p-2.5 border border-slate-100">
      <p className="text-[10px] text-slate-400 uppercase tracking-wide font-medium">{label}</p>
      <p className="text-xs font-semibold text-slate-800 tabular-nums mt-0.5 font-mono">{value}</p>
    </div>
  );
}
