import { Component, Inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatDialogRef, MAT_DIALOG_DATA, MatDialogModule } from '@angular/material/dialog';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { FormsModule } from '@angular/forms';
import { Repository } from '../../../core/services/ganje-api.service';

export interface DeleteDialogData {
  repository: Repository;
}

@Component({
  selector: 'app-repository-delete-dialog',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    MatDialogModule,
    MatButtonModule,
    MatIconModule,
    MatCheckboxModule
  ],
  template: `
    <div class="delete-dialog">
      <div mat-dialog-title class="dialog-title">
        <mat-icon color="warn">warning</mat-icon>
        <span>Delete Repository</span>
      </div>
      
      <div mat-dialog-content class="dialog-content">
        <p class="warning-text">
          Are you sure you want to delete the repository <strong>"{{ data.repository.name }}"</strong>?
        </p>
        
        <div class="repository-info">
          <div class="info-item">
            <strong>Type:</strong> {{ data.repository.type | titlecase }}
          </div>
          <div class="info-item">
            <strong>Artifact Type:</strong> {{ data.repository.artifact_type | uppercase }}
          </div>
          <div class="info-item">
            <strong>Total Artifacts:</strong> {{ data.repository.total_artifacts | number }}
          </div>
          <div class="info-item">
            <strong>Total Size:</strong> {{ formatBytes(data.repository.total_size) }}
          </div>
        </div>
        
        <div class="warning-box">
          <mat-icon>info</mat-icon>
          <div>
            <p><strong>This action cannot be undone.</strong></p>
            <p>All artifacts and metadata in this repository will be permanently deleted.</p>
          </div>
        </div>
        
        <div class="force-option" *ngIf="data.repository.total_artifacts > 0">
          <mat-checkbox [(ngModel)]="forceDelete">
            Force delete (ignore artifacts in use)
          </mat-checkbox>
          <p class="force-hint">
            Check this option if you want to delete the repository even if some artifacts might be in use.
          </p>
        </div>
      </div>
      
      <div mat-dialog-actions class="dialog-actions">
        <button mat-button (click)="onCancel()">Cancel</button>
        <button mat-raised-button color="warn" (click)="onConfirm()">
          <mat-icon>delete</mat-icon>
          Delete Repository
        </button>
      </div>
    </div>
  `,
  styles: [`
    .delete-dialog {
      min-width: 500px;
      max-width: 600px;
    }

    .dialog-title {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      color: #d32f2f;
      font-weight: 500;
    }

    .dialog-content {
      padding: 1rem 0;
    }

    .warning-text {
      font-size: 1.1rem;
      margin-bottom: 1.5rem;
    }

    .repository-info {
      background: #f5f5f5;
      padding: 1rem;
      border-radius: 4px;
      margin-bottom: 1.5rem;
    }

    .info-item {
      margin-bottom: 0.5rem;
    }

    .info-item:last-child {
      margin-bottom: 0;
    }

    .warning-box {
      display: flex;
      gap: 0.75rem;
      padding: 1rem;
      background: #fff3cd;
      border: 1px solid #ffeaa7;
      border-radius: 4px;
      margin-bottom: 1.5rem;
    }

    .warning-box mat-icon {
      color: #856404;
      margin-top: 2px;
    }

    .warning-box p {
      margin: 0 0 0.5rem 0;
    }

    .warning-box p:last-child {
      margin-bottom: 0;
    }

    .force-option {
      margin-top: 1rem;
    }

    .force-hint {
      font-size: 0.875rem;
      color: #666;
      margin: 0.5rem 0 0 24px;
    }

    .dialog-actions {
      display: flex;
      gap: 1rem;
      justify-content: flex-end;
      padding-top: 1rem;
      border-top: 1px solid #e0e0e0;
    }

    @media (max-width: 600px) {
      .delete-dialog {
        min-width: auto;
        width: 100%;
      }
    }
  `]
})
export class RepositoryDeleteDialogComponent {
  forceDelete = false;

  constructor(
    public dialogRef: MatDialogRef<RepositoryDeleteDialogComponent>,
    @Inject(MAT_DIALOG_DATA) public data: DeleteDialogData
  ) {}

  onCancel(): void {
    this.dialogRef.close(false);
  }

  onConfirm(): void {
    this.dialogRef.close({ confirmed: true, force: this.forceDelete });
  }

  formatBytes(bytes: number): string {
    if (bytes === 0) return '0 Bytes';
    
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }
}
