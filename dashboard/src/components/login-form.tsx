"use client";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useAuth } from "@/lib/stores/auth.store";
import { cn } from "@/lib/utils";
import { PiPlanetBold } from "react-icons/pi";
import React from "react";
import { useNavigate } from "react-router";

export function LoginForm({
  className,
  ...props
}: React.ComponentProps<"div">) {
  const { login } = useAuth();
  const navigate = useNavigate();
  const [loading, setLoading] = React.useState(false);

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setLoading(true);

    const form = e.currentTarget;
    const formData = new FormData(form);
    const email = (formData.get("email") || "").toString();
    const password = (formData.get("password") || "").toString();

    try {
      await login(email, password);
      navigate("/");
    } catch (err: unknown) {
      console.error("Login error:", err);
      setLoading(false);
    }
  }

  return (
    <div className={cn("flex flex-col gap-6", className)} {...props}>
      <form onSubmit={handleSubmit}>
        <div className="flex flex-col gap-6">
          <div className="flex flex-col items-center gap-2">
            <div className="flex flex-col items-center gap-2 font-medium">
              <div className="flex size-8 items-center justify-center rounded-md">
                <PiPlanetBold className="size-10" />
              </div>
              <span className="sr-only">Culturae.</span>
            </div>
            <h1 className="text-xl font-bold">Welcome to Culturae.</h1>
          </div>
          <div className="flex flex-col gap-6">
            <div className="grid gap-3">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                name="email"
                type="email"
                placeholder="example@culturae.me"
                required
              />
            </div>
            <div className="grid gap-3">
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                name="password"
                type="password"
                placeholder="*********"
                required
              />
            </div>

            <Button
              type="submit"
              className="w-full bg-primary"
              disabled={loading}
            >
              {loading ? "Signing in..." : "Login"}
            </Button>
          </div>
        </div>
      </form>
    </div>
  );
}
