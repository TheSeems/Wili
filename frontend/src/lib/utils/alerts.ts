import { makeAlert } from "$lib/stores/alerts";
import CheckCircle2Icon from "@lucide/svelte/icons/check-circle-2";
import AlertCircleIcon from "@lucide/svelte/icons/alert-circle";
import XCircleIcon from "@lucide/svelte/icons/x-circle";
import InfoIcon from "@lucide/svelte/icons/info";

export function makeAlertCenter(options: {
  title: string;
  description?: string;
  icon?: any;
  variant?: "default" | "destructive";
  duration?: number;
}) {
  return makeAlert({
    ...options,
    position: "center",
  });
}

export function makeAlertTopRight(options: {
  title: string;
  description?: string;
  icon?: any;
  variant?: "default" | "destructive";
  duration?: number;
}) {
  return makeAlert({
    ...options,
    position: "top-right",
  });
}

export function makeAlertTopCenter(options: {
  title: string;
  description?: string;
  icon?: any;
  variant?: "default" | "destructive";
  duration?: number;
}) {
  return makeAlert({
    ...options,
    position: "top-center",
  });
}

export function makeAlertBottomCenter(options: {
  title: string;
  description?: string;
  icon?: any;
  variant?: "default" | "destructive";
  duration?: number;
}) {
  return makeAlert({
    ...options,
    position: "bottom-center",
  });
}

export function showSuccessAlert(
  title: string,
  description?: string,
  position:
    | "top-center"
    | "top-right"
    | "top-left"
    | "bottom-center"
    | "bottom-right"
    | "bottom-left"
    | "center" = "center"
) {
  return makeAlert({
    title,
    description,
    icon: CheckCircle2Icon,
    variant: "default",
    duration: 5000,
    position,
  });
}

export function showErrorAlert(
  title: string,
  description?: string,
  position:
    | "top-center"
    | "top-right"
    | "top-left"
    | "bottom-center"
    | "bottom-right"
    | "bottom-left"
    | "center" = "center"
) {
  return makeAlert({
    title,
    description,
    icon: XCircleIcon,
    variant: "destructive",
    duration: 7000,
    position,
  });
}

export function showWarningAlert(
  title: string,
  description?: string,
  position:
    | "top-center"
    | "top-right"
    | "top-left"
    | "bottom-center"
    | "bottom-right"
    | "bottom-left"
    | "center" = "center"
) {
  return makeAlert({
    title,
    description,
    icon: AlertCircleIcon,
    variant: "default",
    duration: 6000,
    position,
  });
}

export function showInfoAlert(
  title: string,
  description?: string,
  position:
    | "top-center"
    | "top-right"
    | "top-left"
    | "bottom-center"
    | "bottom-right"
    | "bottom-left"
    | "center" = "center"
) {
  return makeAlert({
    title,
    description,
    icon: InfoIcon,
    variant: "default",
    duration: 5000,
    position,
  });
}
