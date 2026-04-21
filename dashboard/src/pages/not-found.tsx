import { Button } from "@/components/ui/button";
import { Link } from "react-router";

export default function NotFoundPage() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-4">
      <div className="text-center space-y-4">
        <h1 className="text-9xl font-bold text-primary">404</h1>
        <h2 className="text-2xl font-semibold">This page could not be found</h2>
        <p className="text-muted-foreground text-sm max-w-md">
          The page you are looking for doesn&apos;t exist or has been moved.
        </p>
      </div>
      <div className="flex gap-2">
        <Button asChild variant="default">
          <Link to="/">Go to dashboard</Link>
        </Button>
      </div>
    </div>
  );
}
