import { createContext, useContext, useMemo, useState } from "react";
import type { ReactNode } from "react";
import type { User } from "../types";
import { authApi } from "../api/services";
type AuthValue = {
  user: User | null;
  token: string | null;
  login: (e: string, p: string) => Promise<void>;
  register: (n: string, e: string, p: string) => Promise<void>;
  logout: () => void;
};
const Ctx = createContext<AuthValue | null>(null);
export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState(localStorage.getItem("myfinance_token"));
  const [user, setUser] = useState<User | null>(() => {
    try {
      return JSON.parse(localStorage.getItem("myfinance_user") || "null");
    } catch {
      return null;
    }
  });

  const save = (result: { token: string; user: User }) => {
    localStorage.setItem("myfinance_token", result.token);
    localStorage.setItem("myfinance_user", JSON.stringify(result.user));
    setToken(result.token);
    setUser(result.user);
  };

  const value = useMemo<AuthValue>(
    () => ({
      user,
      token,
      login: async (e, p) => save(await authApi.login(e, p)),
      register: async (n, e, p) => save(await authApi.register(n, e, p)),
      logout: () => {
        localStorage.removeItem("myfinance_token");
        localStorage.removeItem("myfinance_user");
        setToken(null);
        setUser(null);
      },
    }),
    [user, token],
  );

  return <Ctx.Provider value={value}>{children}</Ctx.Provider>;
}
export function useAuth() {
  const v = useContext(Ctx);
  if (!v) throw new Error("useAuth must be used inside AuthProvider");
  return v;
}
