"use client";

import { useState, FormEvent } from "react";

// ── Types ────────────────────────────────────────────────────

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
  requestId?: string;
}

// ── Constants ────────────────────────────────────────────────

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

const TRANSPORT_LABELS: Record<string, string> = {
  minivan: "🚐 Mini Van  (0–100 km · ₹3/km/kg)",
  truck: "🚛 Truck  (100–500 km · ₹2/km/kg)",
  aeroplane: "✈️ Aeroplane  (500+ km · ₹1/km/kg)",
};

// ── Helper ───────────────────────────────────────────────────

function fmt(n: number) {
  return n.toLocaleString("en-IN", { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

// ── Component ────────────────────────────────────────────────

export default function Home() {
  const [form, setForm] = useState({
    sellerId: "",
    productId: "",
    customerId: "",
    deliverySpeed: "standard",
  });

  const [result, setResult] = useState<CalculationResult | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(false);

  const handleChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>
  ) => {
    setForm((f) => ({ ...f, [e.target.name]: e.target.value }));
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);
    setFieldErrors({});
    setResult(null);
    setLoading(true);

    try {
      const res = await fetch(`${API_URL}/api/v1/shipping-charge/calculate`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          sellerId: Number(form.sellerId),
          productId: Number(form.productId),
          customerId: Number(form.customerId),
          deliverySpeed: form.deliverySpeed,
        }),
      });

      const json: ApiResponse = await res.json();

      if (!json.success || !json.data) {
        setError(json.error ?? "Unknown error");
        if (json.fields) setFieldErrors(json.fields);
      } else {
        setResult(json.data);
      }
    } catch {
      setError("Failed to reach the API. Make sure the backend is running.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <main className="min-h-screen bg-gradient-to-br from-orange-50 via-white to-amber-50">
      {/* Header */}
      <header className="bg-white border-b border-gray-100 shadow-sm">
        <div className="max-w-4xl mx-auto px-4 py-4 flex items-center gap-3">
          <span className="text-2xl">📦</span>
          <div>
            <h1 className="text-xl font-bold text-gray-900 leading-none">
              Jambotails
            </h1>
            <p className="text-xs text-gray-500 mt-0.5">
              B2B Shipping Charge Estimator
            </p>
          </div>
        </div>
      </header>

      <div className="max-w-4xl mx-auto px-4 py-10 space-y-8">
        {/* Form card */}
        <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-6 md:p-8">
          <h2 className="text-lg font-semibold text-gray-800 mb-1">
            Calculate Shipping Charge
          </h2>
          <p className="text-sm text-gray-500 mb-6">
            Finds nearest warehouse to your seller, then estimates delivery cost
            to your Kirana store.
          </p>

          <form onSubmit={handleSubmit} className="space-y-5">
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
              {/* Seller ID */}
              <div>
                <label className="text-sm font-medium text-gray-700 block mb-1">
                  Seller ID
                </label>
                <input
                  type="number"
                  name="sellerId"
                  value={form.sellerId}
                  onChange={handleChange}
                  placeholder="e.g. 1"
                  min={1}
                  required
                  className="w-full border border-gray-200 rounded-lg px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400 focus:border-transparent transition"
                />
                {fieldErrors.SellerID && (
                  <p className="text-xs text-red-500 mt-1">{fieldErrors.SellerID}</p>
                )}
              </div>

              {/* Product ID */}
              <div>
                <label className="text-sm font-medium text-gray-700 block mb-1">
                  Product ID
                </label>
                <input
                  type="number"
                  name="productId"
                  value={form.productId}
                  onChange={handleChange}
                  placeholder="e.g. 1"
                  min={1}
                  required
                  className="w-full border border-gray-200 rounded-lg px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400 focus:border-transparent transition"
                />
                {fieldErrors.ProductID && (
                  <p className="text-xs text-red-500 mt-1">{fieldErrors.ProductID}</p>
                )}
              </div>

              {/* Customer ID */}
              <div>
                <label className="text-sm font-medium text-gray-700 block mb-1">
                  Customer ID
                </label>
                <input
                  type="number"
                  name="customerId"
                  value={form.customerId}
                  onChange={handleChange}
                  placeholder="e.g. 1"
                  min={1}
                  required
                  className="w-full border border-gray-200 rounded-lg px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400 focus:border-transparent transition"
                />
                {fieldErrors.CustomerID && (
                  <p className="text-xs text-red-500 mt-1">
                    {fieldErrors.CustomerID}
                  </p>
                )}
              </div>
            </div>

            {/* Delivery speed */}
            <div className="sm:w-1/3">
              <label className="text-sm font-medium text-gray-700 block mb-1">
                Delivery Speed
              </label>
              <select
                name="deliverySpeed"
                value={form.deliverySpeed}
                onChange={handleChange}
                className="w-full border border-gray-200 rounded-lg px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400 focus:border-transparent transition bg-white"
              >
                <option value="standard">🐢 Standard</option>
                <option value="express">⚡ Express</option>
              </select>
            </div>

            <button
              type="submit"
              disabled={loading}
              className="w-full sm:w-auto bg-orange-500 hover:bg-orange-600 disabled:bg-orange-300 text-white font-semibold px-8 py-2.5 rounded-lg text-sm transition shadow-sm"
            >
              {loading ? (
                <span className="flex items-center gap-2">
                  <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24" fill="none">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8z" />
                  </svg>
                  Calculating…
                </span>
              ) : (
                "Calculate Shipping →"
              )}
            </button>
          </form>
        </div>

        {/* Error */}
        {error && (
          <div className="bg-red-50 border border-red-200 rounded-xl px-5 py-4 text-sm text-red-700">
            <strong>Error:</strong> {error}
          </div>
        )}

        {/* Results */}
        {result && (
          <div className="space-y-4 animate-in fade-in duration-300">
            {/* Total charge */}
            <div className="bg-gradient-to-r from-orange-500 to-amber-500 rounded-2xl p-6 text-white shadow-md">
              <p className="text-sm font-medium text-orange-100">
                Total Shipping Charge
              </p>
              <p className="text-5xl font-bold mt-1">
                ₹{fmt(result.shippingCharge)}
              </p>
              <p className="text-sm mt-2 text-orange-100">
                {form.deliverySpeed === "express" ? "⚡ Express" : "🐢 Standard"}{" "}
                delivery ·{" "}
                {TRANSPORT_LABELS[result.breakdown.transportMode] ??
                  result.breakdown.transportMode}
              </p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {/* Charge breakdown */}
              <div className="bg-white rounded-2xl border border-gray-100 shadow-sm p-5">
                <h3 className="text-sm font-semibold text-gray-700 mb-3 uppercase tracking-wide">
                  Charge Breakdown
                </h3>
                <table className="w-full text-sm">
                  <tbody className="divide-y divide-gray-50">
                    <Row
                      label="Distance"
                      value={`${fmt(result.breakdown.distanceKm)} km`}
                    />
                    <Row
                      label="Billable Weight"
                      value={`${fmt(result.breakdown.billableWeightKg)} kg`}
                    />
                    <Row
                      label="Rate"
                      value={`₹${result.breakdown.ratePerKmPerKg}/km/kg`}
                    />
                    <Row
                      label="Base Courier Charge"
                      value={`₹${fmt(result.breakdown.baseCourierCharge)}`}
                    />
                    <Row
                      label="Distance Charge"
                      value={`₹${fmt(result.breakdown.distanceCharge)}`}
                    />
                    {result.breakdown.expressCharge > 0 && (
                      <Row
                        label="Express Surcharge"
                        value={`₹${fmt(result.breakdown.expressCharge)}`}
                        highlight
                      />
                    )}
                    <Row
                      label="Total"
                      value={`₹${fmt(result.breakdown.totalCharge)}`}
                      bold
                    />
                  </tbody>
                </table>
              </div>

              {/* Nearest warehouse */}
              <div className="bg-white rounded-2xl border border-gray-100 shadow-sm p-5">
                <h3 className="text-sm font-semibold text-gray-700 mb-3 uppercase tracking-wide">
                  Nearest Warehouse
                </h3>
                <div className="space-y-3">
                  <div>
                    <p className="text-xs text-gray-400">Warehouse Name</p>
                    <p className="text-sm font-semibold text-gray-800">
                      {result.nearestWarehouse.warehouseName}
                    </p>
                  </div>
                  <div>
                    <p className="text-xs text-gray-400">Distance from Seller</p>
                    <p className="text-sm font-semibold text-gray-800">
                      {fmt(result.nearestWarehouse.distanceKm)} km
                    </p>
                  </div>
                  <div>
                    <p className="text-xs text-gray-400">Coordinates</p>
                    <p className="text-sm text-gray-700">
                      {result.nearestWarehouse.warehouseLocation.lat.toFixed(4)}°N,{" "}
                      {result.nearestWarehouse.warehouseLocation.lng.toFixed(4)}°E
                    </p>
                  </div>
                  <a
                    href={`https://www.google.com/maps?q=${result.nearestWarehouse.warehouseLocation.lat},${result.nearestWarehouse.warehouseLocation.lng}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center gap-1.5 text-xs text-orange-600 hover:text-orange-700 font-medium mt-1"
                  >
                    View on Maps ↗
                  </a>
                </div>
              </div>
            </div>

            {/* Formula note */}
            <p className="text-xs text-gray-400 text-center">
              Billable weight = max(actual weight, volumetric weight · L×W×H÷5000) · 
              Total = base + distance charge{result.breakdown.expressCharge > 0 ? " + express surcharge" : ""}
            </p>
          </div>
        )}
      </div>
    </main>
  );
}

// ── Sub-components ───────────────────────────────────────────

function Row({
  label,
  value,
  bold,
  highlight,
}: {
  label: string;
  value: string;
  bold?: boolean;
  highlight?: boolean;
}) {
  return (
    <tr className="py-1.5">
      <td
        className={`py-1.5 text-gray-500 ${bold ? "font-semibold text-gray-800" : ""}`}
      >
        {label}
      </td>
      <td
        className={`py-1.5 text-right tabular-nums ${bold ? "font-bold text-gray-900" : ""} ${highlight ? "text-orange-600 font-medium" : ""}`}
      >
        {value}
      </td>
    </tr>
  );
}
