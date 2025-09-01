import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSnackBar } from '@angular/material/snack-bar';
import { GanjeApiService } from '../../../core/services/ganje-api.service';
import { ErrorHandlerService } from '../../../core/services/error-handler.service';
import { LoadingService } from '../../../core/services/loading.service';
import { LoadingSpinnerComponent } from '../../../shared/components/loading-spinner/loading-spinner.component';

@Component({
  selector: 'app-repository-create',
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
    MatProgressSpinnerModule,
    LoadingSpinnerComponent
  ],
  template: `
    <div class="create-container">
      <mat-card>
        <mat-card-header>
          <mat-card-title>Create New Repository</mat-card-title>
          <mat-card-subtitle>Configure a new artifact repository</mat-card-subtitle>
        </mat-card-header>
        
        <mat-card-content>
          <form [formGroup]="repositoryForm" (ngSubmit)="onSubmit()">
            <div class="form-row">
              <mat-form-field appearance="outline" class="full-width">
                <mat-label>Repository Name</mat-label>
                <input matInput formControlName="name" placeholder="my-repository">
                <mat-error *ngIf="repositoryForm.get('name')?.hasError('required')">
                  Repository name is required
                </mat-error>
                <mat-error *ngIf="repositoryForm.get('name')?.hasError('pattern')">
                  Name must contain only lowercase letters, numbers, and hyphens
                </mat-error>
              </mat-form-field>
            </div>

            <div class="form-row">
              <mat-form-field appearance="outline" class="half-width">
                <mat-label>Repository Type</mat-label>
                <mat-select formControlName="type">
                  <mat-option value="local">Local</mat-option>
                  <mat-option value="remote">Remote</mat-option>
                  <mat-option value="virtual">Virtual</mat-option>
                </mat-select>
                <mat-error *ngIf="repositoryForm.get('type')?.hasError('required')">
                  Repository type is required
                </mat-error>
              </mat-form-field>

              <mat-form-field appearance="outline" class="half-width">
                <mat-label>Artifact Type</mat-label>
                <mat-select formControlName="artifact_type">
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
                <mat-error *ngIf="repositoryForm.get('artifact_type')?.hasError('required')">
                  Artifact type is required
                </mat-error>
              </mat-form-field>
            </div>

            <div class="form-row" *ngIf="repositoryForm.get('type')?.value === 'remote'">
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

            <div class="form-row" *ngIf="repositoryForm.get('type')?.value === 'local'">
              <mat-checkbox formControlName="enable_indexing">
                Enable automatic indexing
              </mat-checkbox>
            </div>

            <div class="form-actions">
              <button mat-button type="button" (click)="onCancel()">Cancel</button>
              <button mat-raised-button color="primary" type="submit" 
                      [disabled]="repositoryForm.invalid || loading">
                <mat-spinner diameter="20" *ngIf="loading"></mat-spinner>
                {{ loading ? 'Creating...' : 'Create Repository' }}
              </button>
            </div>
          </form>
        </mat-card-content>
      </mat-card>
    </div>
  `,
  styles: [`
    .create-container {
      padding: 2rem;
      max-width: 800px;
      margin: 0 auto;
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
export class RepositoryCreateComponent implements OnInit {
  repositoryForm!: FormGroup;
  loading = false;

  constructor(
    private fb: FormBuilder,
    private ganjeApiService: GanjeApiService,
    private router: Router,
    private snackBar: MatSnackBar,
    private errorHandler: ErrorHandlerService,
    private loadingService: LoadingService
  ) {}

  ngOnInit() {
    this.initializeForm();
  }

  private initializeForm() {
    this.repositoryForm = this.fb.group({
      name: ['', [Validators.required, Validators.pattern(/^[a-z0-9-]+$/)]],
      type: ['local', Validators.required],
      artifact_type: ['', Validators.required],
      url: [''],
      description: [''],
      public: [false],
      enable_indexing: [true]
    });

    // Add conditional validation for remote repositories
    this.repositoryForm.get('type')?.valueChanges.subscribe(type => {
      const urlControl = this.repositoryForm.get('url');
      if (type === 'remote') {
        urlControl?.setValidators([Validators.required, Validators.pattern(/^https?:\/\/.+/)]);
      } else {
        urlControl?.clearValidators();
      }
      urlControl?.updateValueAndValidity();
    });
  }

  async onSubmit() {
    if (this.repositoryForm.invalid) {
      this.markFormGroupTouched();
      return;
    }

    const loadingKey = 'repository-create';
    this.loading = true;
    this.loadingService.setLoading(loadingKey, true);
    
    try {
      const formValue = this.repositoryForm.value;
      const repository = {
        name: formValue.name,
        type: formValue.type,
        artifact_type: formValue.artifact_type,
        config: {
          url: formValue.url || undefined,
          description: formValue.description || undefined,
          public: formValue.public,
          enable_indexing: formValue.enable_indexing
        }
      };

      await this.ganjeApiService.createRepository(repository).toPromise();
      this.errorHandler.handleSuccess('Repository created successfully!');
      this.router.navigate(['/repositories']);
    } catch (error: any) {
      this.errorHandler.handleError(error, 'Creating repository');
    } finally {
      this.loading = false;
      this.loadingService.setLoading(loadingKey, false);
    }
  }

  onCancel() {
    this.router.navigate(['/repositories']);
  }

  private markFormGroupTouched() {
    Object.keys(this.repositoryForm.controls).forEach(key => {
      const control = this.repositoryForm.get(key);
      control?.markAsTouched();
    });
  }
}
