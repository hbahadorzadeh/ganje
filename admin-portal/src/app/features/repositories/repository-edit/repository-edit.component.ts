import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { Router, ActivatedRoute } from '@angular/router';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSnackBar } from '@angular/material/snack-bar';
import { GanjeApiService, Repository } from '../../../core/services/ganje-api.service';

@Component({
  selector: 'app-repository-edit',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatCardModule,
    MatButtonModule,
    MatFormFieldModule,
    MatInputModule,
    MatSelectModule,
    MatCheckboxModule,
    MatProgressSpinnerModule
  ],
  template: `
    <div class="edit-container">
      <mat-card>
        <mat-card-header>
          <mat-card-title>Edit Repository</mat-card-title>
          <mat-card-subtitle>Modify repository configuration</mat-card-subtitle>
        </mat-card-header>
        
        <mat-card-content>
          <div *ngIf="loading" class="loading-container">
            <mat-spinner></mat-spinner>
            <p>Loading repository...</p>
          </div>

          <form *ngIf="!loading && repositoryForm" [formGroup]="repositoryForm" (ngSubmit)="onSubmit()">
            <div class="form-row">
              <mat-form-field appearance="outline" class="full-width">
                <mat-label>Repository Name</mat-label>
                <input matInput formControlName="name" [readonly]="true">
                <mat-hint>Repository name cannot be changed after creation</mat-hint>
              </mat-form-field>
            </div>

            <div class="form-row">
              <mat-form-field appearance="outline" class="half-width">
                <mat-label>Repository Type</mat-label>
                <mat-select formControlName="type" [disabled]="true">
                  <mat-option value="local">Local</mat-option>
                  <mat-option value="remote">Remote</mat-option>
                  <mat-option value="virtual">Virtual</mat-option>
                </mat-select>
                <mat-hint>Repository type cannot be changed</mat-hint>
              </mat-form-field>

              <mat-form-field appearance="outline" class="half-width">
                <mat-label>Artifact Type</mat-label>
                <mat-select formControlName="artifact_type" [disabled]="true">
                  <mat-option value="maven">Maven</mat-option>
                  <mat-option value="npm">NPM</mat-option>
                  <mat-option value="docker">Docker</mat-option>
                  <mat-option value="pypi">PyPI</mat-option>
                  <mat-option value="helm">Helm</mat-option>
                  <mat-option value="go">Go Modules</mat-option>
                  <mat-option value="cargo">Cargo</mat-option>
                  <mat-option value="nuget">NuGet</mat-option>
                  <mat-option value="rubygems">RubyGems</mat-option>
                  <mat-option value="generic">Generic</mat-option>
                </mat-select>
                <mat-hint>Artifact type cannot be changed</mat-hint>
              </mat-form-field>
            </div>

            <div class="form-row" *ngIf="repository?.type === 'remote'">
              <mat-form-field appearance="outline" class="full-width">
                <mat-label>Remote URL</mat-label>
                <input matInput formControlName="url" placeholder="https://repo1.maven.org/maven2/">
                <mat-error *ngIf="repositoryForm.get('url')?.hasError('required')">
                  Remote URL is required for remote repositories
                </mat-error>
                <mat-error *ngIf="repositoryForm.get('url')?.hasError('pattern')">
                  Please enter a valid URL
                </mat-error>
              </mat-form-field>
            </div>

            <div class="form-row">
              <mat-form-field appearance="outline" class="full-width">
                <mat-label>Description</mat-label>
                <textarea matInput formControlName="description" rows="3" 
                         placeholder="Optional description for this repository"></textarea>
              </mat-form-field>
            </div>

            <div class="form-row">
              <mat-checkbox formControlName="public">
                Make this repository publicly accessible
              </mat-checkbox>
            </div>

            <div class="form-row" *ngIf="repository?.type === 'local'">
              <mat-checkbox formControlName="enable_indexing">
                Enable automatic indexing
              </mat-checkbox>
            </div>

            <div class="form-actions">
              <button mat-button type="button" (click)="onCancel()">Cancel</button>
              <button mat-raised-button color="primary" type="submit" 
                      [disabled]="repositoryForm.invalid || saving">
                <mat-spinner diameter="20" *ngIf="saving"></mat-spinner>
                {{ saving ? 'Saving...' : 'Save Changes' }}
              </button>
            </div>
          </form>
        </mat-card-content>
      </mat-card>
    </div>
  `,
  styles: [`
    .edit-container {
      padding: 2rem;
      max-width: 800px;
      margin: 0 auto;
    }

    .loading-container {
      display: flex;
      flex-direction: column;
      align-items: center;
      padding: 2rem;
      gap: 1rem;
    }

    .form-row {
      display: flex;
      gap: 1rem;
      margin-bottom: 1rem;
    }

    .full-width {
      width: 100%;
    }

    .half-width {
      width: calc(50% - 0.5rem);
    }

    .form-actions {
      display: flex;
      gap: 1rem;
      justify-content: flex-end;
      margin-top: 2rem;
      padding-top: 1rem;
      border-top: 1px solid #e0e0e0;
    }

    mat-spinner {
      margin-right: 8px;
    }

    @media (max-width: 768px) {
      .form-row {
        flex-direction: column;
      }
      
      .half-width {
        width: 100%;
      }
    }
  `]
})
export class RepositoryEditComponent implements OnInit {
  repositoryForm?: FormGroup;
  repository?: Repository;
  loading = true;
  saving = false;
  repositoryName?: string;

  constructor(
    private fb: FormBuilder,
    private ganjeApiService: GanjeApiService,
    private router: Router,
    private route: ActivatedRoute,
    private snackBar: MatSnackBar
  ) {}

  ngOnInit() {
    this.repositoryName = this.route.snapshot.paramMap.get('name') || '';
    this.loadRepository();
  }

  private async loadRepository() {
    if (!this.repositoryName) {
      this.snackBar.open('Repository name not provided', 'Close', { duration: 3000 });
      this.router.navigate(['/repositories']);
      return;
    }

    try {
      this.repository = await this.ganjeApiService.getRepository(this.repositoryName).toPromise();
      this.initializeForm();
    } catch (error: any) {
      this.snackBar.open(
        error.error?.message || 'Failed to load repository',
        'Close',
        { duration: 5000 }
      );
      this.router.navigate(['/repositories']);
    } finally {
      this.loading = false;
    }
  }

  private initializeForm() {
    if (!this.repository) return;

    const config = this.repository.config || {};
    
    this.repositoryForm = this.fb.group({
      name: [{ value: this.repository.name, disabled: true }],
      type: [{ value: this.repository.type, disabled: true }],
      artifact_type: [{ value: this.repository.artifact_type, disabled: true }],
      url: [config.url || ''],
      description: [config.description || ''],
      public: [config.public || false],
      enable_indexing: [config.enable_indexing !== false] // default to true
    });

    // Add conditional validation for remote repositories
    if (this.repository.type === 'remote') {
      const urlControl = this.repositoryForm.get('url');
      urlControl?.setValidators([Validators.required, Validators.pattern(/^https?:\/\/.+/)]);
      urlControl?.updateValueAndValidity();
    }
  }

  async onSubmit() {
    if (!this.repositoryForm || this.repositoryForm.invalid || !this.repository) {
      this.markFormGroupTouched();
      return;
    }

    this.saving = true;
    try {
      const formValue = this.repositoryForm.value;
      const updatedRepository = {
        ...this.repository,
        config: {
          ...this.repository.config,
          url: formValue.url || undefined,
          description: formValue.description || undefined,
          public: formValue.public,
          enable_indexing: formValue.enable_indexing
        }
      };

      await this.ganjeApiService.updateRepository(this.repository.name, updatedRepository).toPromise();
      this.snackBar.open('Repository updated successfully!', 'Close', { duration: 3000 });
      this.router.navigate(['/repositories']);
    } catch (error: any) {
      this.snackBar.open(
        error.error?.message || 'Failed to update repository. Please try again.',
        'Close',
        { duration: 5000 }
      );
    } finally {
      this.saving = false;
    }
  }

  onCancel() {
    this.router.navigate(['/repositories']);
  }

  private markFormGroupTouched() {
    if (!this.repositoryForm) return;
    
    Object.keys(this.repositoryForm.controls).forEach(key => {
      const control = this.repositoryForm!.get(key);
      control?.markAsTouched();
    });
  }
}
