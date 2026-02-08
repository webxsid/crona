import { useStdout } from "ink";
import { useEffect, useState } from "react";

export function useStdoutDimensions() {
  const { stdout } = useStdout();

  const get = () => ({
    columns: stdout.columns ?? 80,
    rows: stdout.rows ?? 24,
  });

  const [dims, setDims] = useState(get);

  useEffect(() => {
    if (!stdout) return;

    const onResize = () => setDims(get());
    stdout.on("resize", onResize);

    return () => {
      stdout.off("resize", onResize);
    };
  }, [stdout]);

  return dims;
}
