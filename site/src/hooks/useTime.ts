/**
 * Remaining things to do:
 * @todo Figure out how DayJS values are actually structured
 * @todo Figure out how to turn default TTransformed type param from type
 *       undefined to type never without breaking everything (undefined should
 *       be a valid return type for the transformed function, but the current
 *       setup treats it as a cue to ignore the value, and go with the base
 *       date value)
 */
import { useEffect, useState } from "react";
import { useEffectEvent } from "./hookPolyfills";

type DayJsValue = {
  hah: "yeah";
};

type TimeValueFormat = "date" | "dayjs" | undefined;
type TimeValue<T extends TimeValueFormat> = T extends "dayjs" | undefined
  ? DayJsValue
  : T extends "date"
    ? Date
    : never;

type Transform<TFormat extends TimeValueFormat, TTransformed> = (
  currentDate: TimeValue<TFormat>,
  // Adding Awaited here is a "hack" to ensure that async functions can't
  // accidentally be passed in as transform values; synchronous functions only!
) => Awaited<TTransformed>;

type UseRelativeTimeReturnValue<
  TFormat extends TimeValueFormat,
  TTransformed,
> = TTransformed extends undefined
  ? TimeValue<TFormat>
  : ReturnType<Transform<TFormat, TTransformed>>;

export type ConfigOptions<
  TFormat extends TimeValueFormat,
  TTransformed,
> = Readonly<{
  /**
   * Determines how often the hook will re-render with a new value. The value is
   * allowed to change on re-render, but doing so will always restart the
   * refresh interval from scratch.
   *
   * Defaults to 1000 milliseconds/1 second if not specified.
   */
  refreshIntervalMs?: number;

  /**
   * Determines whether the hook will keep re-rendering on the given refresh
   * interval. If set to false, this hook will never re-render.
   *
   * Defaults to true if not specified.
   */
  enabled?: boolean;

  /**
   * The type of "base value" to store in state on each re-render. Can be
   * transformed into other values via the config's transform property.
   *
   * Defaults to type "dayjs" if not specified.
   */
  rawTimeFormat?: TFormat;

  /**
   * A transformation callback for taking the newest time value, and turning it
   * into something else.
   *
   * This function can be non-deterministic, but it should not make have logic
   * with heavy side effects (network requests, etc.) and should not be
   * promise-based. The function will be evaluated eagerly on the mounting
   * render, but will be re-evaluated via effects for all re-renders.
   *
   * If not specified, the hook will return out the current time with zero
   * changes, as specified by the rawTimeFormat property.
   */
  transform?: Transform<TFormat, TTransformed>;
}>;

export function useTime<
  TFormat extends TimeValueFormat = undefined,
  TTransformed = undefined,
>(
  config?: ConfigOptions<TFormat, TTransformed>,
): UseRelativeTimeReturnValue<TFormat, TTransformed> {
  const {
    transform,
    refreshIntervalMs = 1_000,
    enabled = true,
    rawTimeFormat = "dayjs",
  } = config ?? {};

  const createFormattedTimeValue = useEffectEvent(() => {
    const newTimeValue =
      rawTimeFormat === "dayjs" ? { hah: "yeah" } : new Date();

    const output =
      transform?.(newTimeValue as TimeValue<TFormat>) ?? newTimeValue;

    return output as UseRelativeTimeReturnValue<TFormat, TTransformed>;
  });

  // Have to break the function purity rules on the mounting render by having
  // a non-deterministic value be initialized in state, just so there's always a
  // value available, but all re-renders go through useEffect and remain pure
  const [timeValue, setTimeValue] = useState(createFormattedTimeValue);
  useEffect(() => {
    if (!enabled) {
      return undefined;
    }

    const intervalId = window.setInterval(() => {
      setTimeValue(createFormattedTimeValue());
    }, refreshIntervalMs);

    return () => window.clearInterval(intervalId);
  }, [createFormattedTimeValue, refreshIntervalMs, enabled]);

  return timeValue;
}

export function Test() {
  const [needRefresh, setNeedRefresh] = useState(true);
  const date = useTime({
    refreshIntervalMs: 5_000,
    enabled: needRefresh,
    rawTimeFormat: "date",
    transform: () => true as const,
  });

  return { setNeedRefresh, date } as const;
}
