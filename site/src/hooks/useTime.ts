import { useEffect, useState } from "react";
import { useEffectEvent } from "./hookPolyfills";
import dayjs, { type Dayjs } from "dayjs";

type TimeValueFormat = "date" | "dayjs";
type TimeValue<T extends TimeValueFormat> = T extends "dayjs"
  ? Dayjs
  : T extends "date"
    ? Date
    : never;

type Transform<TFormat extends TimeValueFormat, TTransformed> = (
  currentDate: TimeValue<TFormat>,
  // Adding Awaited here is a "hack" to ensure that async functions can't
  // accidentally be passed in as transform values; synchronous functions only!
) => Awaited<TTransformed>;

type UseTimeReturnValue<TFormat extends TimeValueFormat, TTransformed> =
  // Have to use tuples for comparison to avoid type contravariance issues
  [TTransformed] extends [never]
    ? TimeValue<TFormat>
    : ReturnType<Transform<TFormat, TTransformed>>;

export type UseTimeConfig<
  TFormat extends TimeValueFormat,
  TTransformed = unknown,
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
   * interval. This hook will never cause itself to re-render as long as paused
   * is true.
   *
   * Defaults to false if not specified.
   */
  paused?: boolean;

  /**
   * The type of the "base time value" to use for calculating useTime's state.
   * Can be transformed into other values via the config's transform property.
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
  TFormat extends TimeValueFormat = "dayjs",
  TTransformed = never,
>(
  config?: UseTimeConfig<TFormat, TTransformed>,
): UseTimeReturnValue<TFormat, TTransformed> {
  const {
    transform,
    refreshIntervalMs = 1_000,
    paused = false,
    rawTimeFormat = "dayjs",
  } = config ?? {};

  // Not a fan of the type assertions, but the alternative would involve
  // jumping through a bunch of hoops. Not worth it for so few lines of code
  const createFormattedTimeValue = useEffectEvent(() => {
    type Return = UseTimeReturnValue<TFormat, TTransformed>;
    const newTimeValue = rawTimeFormat === "dayjs" ? dayjs() : new Date();
    if (transform === undefined) {
      return newTimeValue as Return;
    }

    return transform(newTimeValue as TimeValue<TFormat>) as Return;
  });

  // Have to break the function purity rules on the mounting render by having
  // a non-deterministic value be initialized in state, just so there's always a
  // value available. But all re-renders go through useEffect, keeping the
  // render logic 100% pure afterwards
  const [timeValue, setTimeValue] = useState(createFormattedTimeValue);
  useEffect(() => {
    if (paused) {
      return;
    }

    const intervalId = window.setInterval(() => {
      setTimeValue(createFormattedTimeValue());
    }, refreshIntervalMs);

    return () => window.clearInterval(intervalId);
  }, [createFormattedTimeValue, refreshIntervalMs, paused]);

  return timeValue;
}
