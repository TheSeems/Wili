import { writable } from "svelte/store";

export interface AlertItem {
  id: string;
  title: string;
  description?: string;
  variant?: "default" | "destructive";
  icon?: any;
  duration?: number;
  position?:
    | "top-center"
    | "top-right"
    | "top-left"
    | "bottom-center"
    | "bottom-right"
    | "bottom-left"
    | "center";
}

export const alertsStore = writable<AlertItem[]>([]);

export function makeAlert(alert: Omit<AlertItem, "id">) {
  const id = crypto.randomUUID();
  const duration = alert.duration ?? 5000;
  const position = alert.position ?? "top-center";

  const alertItem: AlertItem = {
    id,
    position,
    ...alert,
  };

  alertsStore.update((alerts) => [...alerts, alertItem]);

  setTimeout(() => {
    removeAlert(id);
  }, duration);

  return id;
}

export function removeAlert(id: string) {
  alertsStore.update((alerts) => alerts.filter((alert) => alert.id !== id));
}
