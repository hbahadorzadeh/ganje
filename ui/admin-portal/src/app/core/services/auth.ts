import { Injectable } from '@angular/core';
import { Router } from '@angular/router';
import { OAuthService, AuthConfig } from 'angular-oauth2-oidc';

const authConfig: AuthConfig = {
  // TODO: replace placeholders
  issuer: 'https://your-issuer.example.com/realms/yourrealm',
  clientId: 'admin-portal-client',
  redirectUri: window.location.origin + '/auth/callback',
  postLogoutRedirectUri: window.location.origin + '/',
  responseType: 'code',
  scope: 'openid profile email',
  showDebugInformation: false,
  requireHttps: true,
  useSilentRefresh: false,
};

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  constructor(private oauth: OAuthService, private router: Router) {
    this.oauth.configure(authConfig);
    // Attempt to log in on app start if coming back from IdP
    this.oauth.loadDiscoveryDocumentAndTryLogin();
  }

  login(): void {
    this.oauth.initCodeFlow();
  }

  async processCallback(): Promise<void> {
    // loadDiscoveryDocumentAndTryLogin already processes, but ensure token validity
    if (!this.oauth.hasValidAccessToken()) {
      await this.oauth.loadDiscoveryDocumentAndTryLogin();
    }
  }

  logout(): void {
    this.oauth.logOut();
  }

  isAuthenticated(): boolean {
    return this.oauth.hasValidAccessToken();
  }

  get accessToken(): string | null {
    const t = this.oauth.getAccessToken();
    return t || null;
  }
}
