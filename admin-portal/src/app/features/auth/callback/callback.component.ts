import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { AuthService } from '../../../core/services/auth.service';
import { CommonModule } from '@angular/common';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';

@Component({
  selector: 'app-auth-callback',
  standalone: true,
  imports: [CommonModule, MatProgressSpinnerModule],
  template: `
    <div class="callback-container">
      <mat-spinner></mat-spinner>
      <p>Processing authentication...</p>
      <p *ngIf="error" class="error">{{ error }}</p>
    </div>
  `,
  styles: [`
    .callback-container {
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      height: 100vh;
      gap: 1rem;
    }
    
    .error {
      color: #f44336;
      text-align: center;
    }
  `]
})
export class CallbackComponent implements OnInit {
  error: string | null = null;

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private authService: AuthService
  ) {}

  ngOnInit() {
    this.route.queryParams.subscribe(params => {
      const code = params['code'];
      const state = params['state'];
      const error = params['error'];

      if (error) {
        this.error = `Authentication failed: ${error}`;
        setTimeout(() => this.router.navigate(['/login']), 3000);
        return;
      }

      if (!code) {
        this.error = 'No authorization code received';
        setTimeout(() => this.router.navigate(['/login']), 3000);
        return;
      }

      // Validate state parameter
      const storedState = localStorage.getItem('oauth_state');
      if (state !== storedState) {
        this.error = 'Invalid state parameter';
        setTimeout(() => this.router.navigate(['/login']), 3000);
        return;
      }

      // Clear stored state
      localStorage.removeItem('oauth_state');

      // Exchange code for token
      this.authService.handleAuthCallback(code).subscribe({
        next: (response) => {
          this.authService.setToken(response.token, response.user);
          this.router.navigate(['/dashboard']);
        },
        error: (err) => {
          console.error('Auth callback error:', err);
          this.error = 'Failed to complete authentication';
          setTimeout(() => this.router.navigate(['/login']), 3000);
        }
      });
    });
  }
}
