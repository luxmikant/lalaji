import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Jambotails — Shipping Estimator",
  description: "B2B shipping charge estimator for Kirana stores",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className="min-h-screen bg-gray-50 font-sans antialiased">
        {children}
      </body>
    </html>
  );
}
