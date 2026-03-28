import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Breeze - Partner Dashboard",
  description: "Partner availability and date management",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className="min-h-screen bg-gray-50">
        {children}
      </body>
    </html>
  );
}
