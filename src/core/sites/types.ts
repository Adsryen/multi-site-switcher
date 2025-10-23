import type { Account } from '../storage/accounts.js'

export interface SwitchOptions {
  [key: string]: any
}

export interface SiteAdapter {
  key: string
  name: string
  loginUrl?: string
  matches(url: string): boolean
  login(account: Account, opts?: SwitchOptions): Promise<any>
  logout?(opts?: SwitchOptions): Promise<any>
}
