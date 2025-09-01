import { Component, Input } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatIconModule } from '@angular/material/icon';
import { MatButtonModule } from '@angular/material/button';

@Component({
  selector: 'app-error-display',
  standalone: true,
  imports: [CommonModule, MatIconModule, MatButtonModule],
  template: `
    <div class="error-container" [class.compact]="compact">
      <div class="error-content">
        <mat-icon class="error-icon">{{ icon }}</mat-icon>
        <div class="error-text">
          <h4 *ngIf="title">{{ title }}</h4>
          <p>{{ message }}</p>
        </div>
        <button *ngIf="retryAction" mat-raised-button color="primary" (click)="retryAction()">
          <mat-icon>refresh</mat-icon>
          Retry
        </button>
      </div>
    </div>
  `,
  styles: [`
    .error-container {
      display: flex;
      align-items: center;
      justify-content: center;
      padding: 2rem;
      text-align: center;
    }

    .error-container.compact {
      padding: 1rem;
    }

    .error-content {
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 1rem;
      max-width: 400px;
    }

    .error-icon {
      font-size: 3rem;
      width: 3rem;
      height: 3rem;
      color: var(--color-danger);
    }

    .compact .error-icon {
      font-size: 2rem;
      width: 2rem;
      height: 2rem;
    }

    .error-text h4 {
      margin: 0 0 0.5rem;
      color: var(--text-basic-color);
    }

    .error-text p {
      margin: 0;
      color: var(--text-hint-color);
      line-height: 1.5;
    }

    .compact .error-text {
      text-align: left;
    }

    .compact .error-content {
      flex-direction: row;
      text-align: left;
      max-width: none;
    }
  `]
})
export class ErrorDisplayComponent {
  @Input() message: string = 'An error occurred';
  @Input() title?: string;
  @Input() icon: string = 'error';
  @Input() compact: boolean = false;
  @Input() retryAction?: () => void;
}
