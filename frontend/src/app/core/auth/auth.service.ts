import { inject, Injectable, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface SignUpInput {
  login: string;
  password: string;
  role_id?: number | null;
  full_name?: string | null;
}

export interface SignInInput {
  login: string;
  password: string;
}

export interface AuthResponse {
  token?: string;
  user?: unknown;
}

const TOKEN_KEY = 'auth_token';

@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly http = inject(HttpClient);

  token = signal<string | null>(this.getToken());

  signUp(body: SignUpInput): Observable<AuthResponse> {
    return this.http.post<AuthResponse>(`/auth/sign-up`, body);
  }

  signIn(body: SignInInput): Observable<AuthResponse> {
    return this.http.post<AuthResponse>(`/auth/sign-in`, body);
  }

  saveToken(token: string): void {
    localStorage.setItem(TOKEN_KEY, token);
    this.token.set(token);
  }

  getToken(): string | null {
    return localStorage.getItem(TOKEN_KEY);
  }

  clearToken(): void {
    localStorage.removeItem(TOKEN_KEY);
    this.token.set(null);
  }

  isAuthenticated(): boolean {
    return !!this.getToken();
  }
}
