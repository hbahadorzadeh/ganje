import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule, FormArray, FormControl } from '@angular/forms';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatSelectModule } from '@angular/material/select';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { MatTabsModule } from '@angular/material/tabs';
import { MatExpansionModule } from '@angular/material/expansion';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSnackBar } from '@angular/material/snack-bar';
import { ErrorHandlerService } from '../../../core/services/error-handler.service';
import { LoadingService } from '../../../core/services/loading.service';
import { LoadingSpinnerComponent } from '../../../shared/components/loading-spinner/loading-spinner.component';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatDividerModule } from '@angular/material/divider';

export interface DexConfig {
  issuer: string;
  storage: {
    type: string;
    config: any;
  };
  web: {
    http: string;
    https?: string;
    allowedOrigins?: string[];
  };
  staticClients: StaticClient[];
  connectors: Connector[];
  oauth2?: {
    skipApprovalScreen?: boolean;
    alwaysShowLoginScreen?: boolean;
    passwordConnector?: string;
  };
  enablePasswordDB?: boolean;
  staticPasswords?: StaticPassword[];
}

export interface StaticClient {
  id: string;
  redirectURIs: string[];
  name: string;
  secret?: string;
  public?: boolean;
  trustedPeers?: string[];
}

export interface Connector {
  type: string;
  id: string;
  name: string;
  config: any;
}

export interface StaticPassword {
  email: string;
  hash: string;
  username: string;
  userID: string;
}

@Component({
  selector: 'app-dex-config-list',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatCardModule,
    MatButtonModule,
    MatIconModule,
    MatInputModule,
    MatFormFieldModule,
    MatSelectModule,
    MatCheckboxModule,
    MatTabsModule,
    MatExpansionModule,
    MatProgressSpinnerModule,
    MatTooltipModule,
    MatDividerModule,
    LoadingSpinnerComponent
  ],
  template: `
    <div class="dex-config-container">
      <div class="page-header">
        <div class="header-content">
          <h1>Dex OAuth Configuration</h1>
          <p>Configure authentication providers and OAuth settings</p>
        </div>
        <div class="header-actions">
          <button mat-button (click)="resetToDefaults()" [disabled]="loading">
            <mat-icon>restore</mat-icon>
            Reset to Defaults
          </button>
          <button mat-raised-button color="primary" (click)="saveConfiguration()" [disabled]="configForm.invalid || saving">
            <mat-icon>save</mat-icon>
            {{ saving ? 'Saving...' : 'Save Configuration' }}
          </button>
        </div>
      </div>

      <form [formGroup]="configForm" *ngIf="!loading">
        <mat-tab-group>
          <!-- Basic Configuration -->
          <mat-tab label="Basic Settings">
            <div class="tab-content">
              <mat-card>
                <mat-card-header>
                  <mat-card-title>Core Configuration</mat-card-title>
                  <mat-card-subtitle>Basic Dex server settings</mat-card-subtitle>
                </mat-card-header>
                <mat-card-content>
                  <div class="form-row">
                    <mat-form-field appearance="outline" class="full-width">
                      <mat-label>Issuer URL</mat-label>
                      <input matInput formControlName="issuer" placeholder="https://dex.example.com">
                      <mat-hint>The base URL for Dex. This should be publicly accessible.</mat-hint>
                      <mat-error *ngIf="configForm.get('issuer')?.hasError('required')">
                        Issuer URL is required
                      </mat-error>
                      <mat-error *ngIf="configForm.get('issuer')?.hasError('pattern')">
                        Please enter a valid HTTPS URL
                      </mat-error>
                    </mat-form-field>
                  </div>

                  <div formGroupName="web">
                    <div class="form-row">
                      <mat-form-field appearance="outline" class="half-width">
                        <mat-label>HTTP Listen Address</mat-label>
                        <input matInput formControlName="http" placeholder="0.0.0.0:5556">
                        <mat-hint>Address for HTTP server</mat-hint>
                      </mat-form-field>

                      <mat-form-field appearance="outline" class="half-width">
                        <mat-label>HTTPS Listen Address</mat-label>
                        <input matInput formControlName="https" placeholder="0.0.0.0:5554">
                        <mat-hint>Address for HTTPS server (optional)</mat-hint>
                      </mat-form-field>
                    </div>
                  </div>

                  <div formGroupName="oauth2">
                    <div class="form-row">
                      <mat-checkbox formControlName="skipApprovalScreen">
                        Skip approval screen for trusted clients
                      </mat-checkbox>
                    </div>
                    <div class="form-row">
                      <mat-checkbox formControlName="alwaysShowLoginScreen">
                        Always show login screen
                      </mat-checkbox>
                    </div>
                    <div class="form-row">
                      <mat-checkbox formControlName="enablePasswordDB">
                        Enable password database
                      </mat-checkbox>
                    </div>
                  </div>
                </mat-card-content>
              </mat-card>
            </div>
          </mat-tab>

          <!-- Storage Configuration -->
          <mat-tab label="Storage">
            <div class="tab-content">
              <mat-card>
                <mat-card-header>
                  <mat-card-title>Storage Backend</mat-card-title>
                  <mat-card-subtitle>Configure where Dex stores its data</mat-card-subtitle>
                </mat-card-header>
                <mat-card-content>
                  <div formGroupName="storage">
                    <div class="form-row">
                      <mat-form-field appearance="outline" class="full-width">
                        <mat-label>Storage Type</mat-label>
                        <mat-select formControlName="type" (selectionChange)="onStorageTypeChange($event.value)">
                          <mat-option value="memory">Memory (Development only)</mat-option>
                          <mat-option value="sqlite3">SQLite</mat-option>
                          <mat-option value="postgres">PostgreSQL</mat-option>
                          <mat-option value="mysql">MySQL</mat-option>
                          <mat-option value="kubernetes">Kubernetes CRDs</mat-option>
                          <mat-option value="etcd">etcd</mat-option>
                        </mat-select>
                      </mat-form-field>
                    </div>

                    <!-- Storage-specific configuration -->
                    <div *ngIf="storageType === 'sqlite3'" class="storage-config">
                      <mat-form-field appearance="outline" class="full-width">
                        <mat-label>Database File Path</mat-label>
                        <input matInput [formControl]="getStorageConfigControl('file')" placeholder="/var/dex/dex.db">
                      </mat-form-field>
                    </div>

                    <div *ngIf="storageType === 'postgres' || storageType === 'mysql'" class="storage-config">
                      <div class="form-row">
                        <mat-form-field appearance="outline" class="half-width">
                          <mat-label>Host</mat-label>
                          <input matInput [formControl]="getStorageConfigControl('host')" placeholder="localhost">
                        </mat-form-field>
                        <mat-form-field appearance="outline" class="half-width">
                          <mat-label>Port</mat-label>
                          <input matInput type="number" [formControl]="getStorageConfigControl('port')" 
                                 [placeholder]="storageType === 'postgres' ? '5432' : '3306'">
                        </mat-form-field>
                      </div>
                      <div class="form-row">
                        <mat-form-field appearance="outline" class="half-width">
                          <mat-label>Database Name</mat-label>
                          <input matInput [formControl]="getStorageConfigControl('database')" placeholder="dex">
                        </mat-form-field>
                        <mat-form-field appearance="outline" class="half-width">
                          <mat-label>Username</mat-label>
                          <input matInput [formControl]="getStorageConfigControl('user')" placeholder="dex">
                        </mat-form-field>
                      </div>
                      <div class="form-row">
                        <mat-form-field appearance="outline" class="full-width">
                          <mat-label>Password</mat-label>
                          <input matInput type="password" [formControl]="getStorageConfigControl('password')">
                        </mat-form-field>
                      </div>
                      <div class="form-row">
                        <mat-checkbox [formControl]="getStorageConfigControl('ssl')">
                          Enable SSL
                        </mat-checkbox>
                      </div>
                    </div>
                  </div>
                </mat-card-content>
              </mat-card>
            </div>
          </mat-tab>

          <!-- Static Clients -->
          <mat-tab label="OAuth Clients">
            <div class="tab-content">
              <mat-card>
                <mat-card-header>
                  <div class="card-header-content">
                    <div>
                      <mat-card-title>OAuth2 Clients</mat-card-title>
                      <mat-card-subtitle>Configure applications that can authenticate with Dex</mat-card-subtitle>
                    </div>
                    <button mat-raised-button color="primary" (click)="addStaticClient()">
                      <mat-icon>add</mat-icon>
                      Add Client
                    </button>
                  </div>
                </mat-card-header>
                <mat-card-content>
                  <div formArrayName="staticClients">
                    <mat-expansion-panel *ngFor="let client of staticClientsArray.controls; let i = index" 
                                         [formGroupName]="i" class="client-panel">
                      <mat-expansion-panel-header>
                        <mat-panel-title>
                          {{ client.get('name')?.value || 'New Client' }}
                        </mat-panel-title>
                        <mat-panel-description>
                          ID: {{ client.get('id')?.value || 'Not set' }}
                        </mat-panel-description>
                      </mat-expansion-panel-header>

                      <div class="client-form">
                        <div class="form-row">
                          <mat-form-field appearance="outline" class="half-width">
                            <mat-label>Client ID</mat-label>
                            <input matInput formControlName="id" placeholder="my-app">
                            <mat-error *ngIf="client.get('id')?.hasError('required')">
                              Client ID is required
                            </mat-error>
                          </mat-form-field>
                          <mat-form-field appearance="outline" class="half-width">
                            <mat-label>Client Name</mat-label>
                            <input matInput formControlName="name" placeholder="My Application">
                            <mat-error *ngIf="client.get('name')?.hasError('required')">
                              Client name is required
                            </mat-error>
                          </mat-form-field>
                        </div>

                        <div class="form-row">
                          <mat-form-field appearance="outline" class="full-width">
                            <mat-label>Client Secret</mat-label>
                            <input matInput type="password" formControlName="secret" 
                                   placeholder="Leave empty for public clients">
                            <mat-hint>Required for confidential clients</mat-hint>
                          </mat-form-field>
                        </div>

                        <div class="form-row">
                          <mat-form-field appearance="outline" class="full-width">
                            <mat-label>Redirect URIs (one per line)</mat-label>
                            <textarea matInput formControlName="redirectURIs" rows="3"
                                      placeholder="https://myapp.example.com/callback&#10;http://localhost:8080/callback"></textarea>
                            <mat-hint>Valid redirect URIs for this client</mat-hint>
                          </mat-form-field>
                        </div>

                        <div class="form-row">
                          <mat-checkbox formControlName="public">
                            Public client (no secret required)
                          </mat-checkbox>
                        </div>

                        <div class="client-actions">
                          <button mat-button color="warn" (click)="removeStaticClient(i)">
                            <mat-icon>delete</mat-icon>
                            Remove Client
                          </button>
                        </div>
                      </div>
                    </mat-expansion-panel>
                  </div>

                  <div *ngIf="staticClientsArray.length === 0" class="empty-state">
                    <mat-icon>apps</mat-icon>
                    <p>No OAuth clients configured</p>
                    <button mat-raised-button color="primary" (click)="addStaticClient()">
                      Add Your First Client
                    </button>
                  </div>
                </mat-card-content>
              </mat-card>
            </div>
          </mat-tab>

          <!-- Connectors -->
          <mat-tab label="Identity Providers">
            <div class="tab-content">
              <mat-card>
                <mat-card-header>
                  <div class="card-header-content">
                    <div>
                      <mat-card-title>Identity Providers</mat-card-title>
                      <mat-card-subtitle>Configure external authentication providers</mat-card-subtitle>
                    </div>
                    <button mat-raised-button color="primary" (click)="addConnector()">
                      <mat-icon>add</mat-icon>
                      Add Provider
                    </button>
                  </div>
                </mat-card-header>
                <mat-card-content>
                  <div formArrayName="connectors">
                    <mat-expansion-panel *ngFor="let connector of connectorsArray.controls; let i = index" 
                                         [formGroupName]="i" class="connector-panel">
                      <mat-expansion-panel-header>
                        <mat-panel-title>
                          {{ connector.get('name')?.value || 'New Provider' }}
                        </mat-panel-title>
                        <mat-panel-description>
                          Type: {{ connector.get('type')?.value || 'Not set' }}
                        </mat-panel-description>
                      </mat-expansion-panel-header>

                      <div class="connector-form">
                        <div class="form-row">
                          <mat-form-field appearance="outline" class="third-width">
                            <mat-label>Provider Type</mat-label>
                            <mat-select formControlName="type">
                              <mat-option value="ldap">LDAP</mat-option>
                              <mat-option value="github">GitHub</mat-option>
                              <mat-option value="gitlab">GitLab</mat-option>
                              <mat-option value="google">Google</mat-option>
                              <mat-option value="microsoft">Microsoft</mat-option>
                              <mat-option value="oidc">Generic OIDC</mat-option>
                              <mat-option value="saml">SAML</mat-option>
                            </mat-select>
                          </mat-form-field>
                          <mat-form-field appearance="outline" class="third-width">
                            <mat-label>Provider ID</mat-label>
                            <input matInput formControlName="id" placeholder="ldap">
                          </mat-form-field>
                          <mat-form-field appearance="outline" class="third-width">
                            <mat-label>Display Name</mat-label>
                            <input matInput formControlName="name" placeholder="Corporate LDAP">
                          </mat-form-field>
                        </div>

                        <mat-divider></mat-divider>

                        <div class="connector-config">
                          <h4>Provider Configuration</h4>
                          <p class="config-hint">
                            Configuration varies by provider type. Please refer to Dex documentation for specific settings.
                          </p>
                          
                          <mat-form-field appearance="outline" class="full-width">
                            <mat-label>Configuration (JSON)</mat-label>
                            <textarea matInput formControlName="config" rows="8" 
                                      placeholder='{"clientID": "...", "clientSecret": "...", "redirectURI": "..."}'></textarea>
                            <mat-hint>Provider-specific configuration in JSON format</mat-hint>
                          </mat-form-field>
                        </div>

                        <div class="connector-actions">
                          <button mat-button color="warn" (click)="removeConnector(i)">
                            <mat-icon>delete</mat-icon>
                            Remove Provider
                          </button>
                        </div>
                      </div>
                    </mat-expansion-panel>
                  </div>

                  <div *ngIf="connectorsArray.length === 0" class="empty-state">
                    <mat-icon>account_circle</mat-icon>
                    <p>No identity providers configured</p>
                    <button mat-raised-button color="primary" (click)="addConnector()">
                      Add Your First Provider
                    </button>
                  </div>
                </mat-card-content>
              </mat-card>
            </div>
          </mat-tab>
        </mat-tab-group>
      </form>

      <div *ngIf="loading" class="loading-container">
        <mat-spinner></mat-spinner>
        <p>Loading configuration...</p>
      </div>
    </div>
  `,
  styles: [`
    .dex-config-container {
      padding: 2rem;
      max-width: 1200px;
      margin: 0 auto;
    }

    .page-header {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      margin-bottom: 2rem;
    }

    .header-content h1 {
      margin: 0 0 0.5rem;
    }

    .header-content p {
      margin: 0;
      color: #666;
    }

    .header-actions {
      display: flex;
      gap: 1rem;
    }

    .tab-content {
      padding: 1rem 0;
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

    .third-width {
      width: calc(33.333% - 0.67rem);
    }

    .card-header-content {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      width: 100%;
    }

    .storage-config {
      margin-top: 1rem;
      padding-top: 1rem;
      border-top: 1px solid #e0e0e0;
    }

    .client-panel,
    .connector-panel {
      margin-bottom: 1rem;
    }

    .client-form,
    .connector-form {
      padding: 1rem 0;
    }

    .client-actions,
    .connector-actions {
      display: flex;
      justify-content: flex-end;
      margin-top: 1rem;
      padding-top: 1rem;
      border-top: 1px solid #e0e0e0;
    }

    .connector-config {
      margin: 1rem 0;
    }

    .config-hint {
      color: #666;
      font-size: 0.875rem;
      margin-bottom: 1rem;
    }

    .empty-state {
      text-align: center;
      padding: 3rem;
      color: #666;
    }

    .empty-state mat-icon {
      font-size: 4rem;
      width: 4rem;
      height: 4rem;
      margin-bottom: 1rem;
    }

    .loading-container {
      display: flex;
      flex-direction: column;
      align-items: center;
      padding: 4rem;
      gap: 1rem;
    }

    @media (max-width: 768px) {
      .page-header {
        flex-direction: column;
        gap: 1rem;
      }
      
      .header-actions {
        width: 100%;
        justify-content: flex-end;
      }
      
      .form-row {
        flex-direction: column;
      }
      
      .half-width,
      .third-width {
        width: 100%;
      }
    }
  `]
})
export class DexConfigListComponent implements OnInit {
  configForm!: FormGroup;
  loading = false;
  saving = false;
  storageType = 'memory';

  constructor(
    private fb: FormBuilder,
    private snackBar: MatSnackBar,
    private errorHandler: ErrorHandlerService,
    private loadingService: LoadingService
  ) {}

  ngOnInit() {
    this.initializeForm();
    this.loadConfiguration();
  }

  private initializeForm() {
    this.configForm = this.fb.group({
      issuer: ['', [Validators.required, Validators.pattern(/^https:\/\/.+/)]],
      storage: this.fb.group({
        type: ['memory', Validators.required],
        config: this.fb.group({})
      }),
      web: this.fb.group({
        http: ['0.0.0.0:5556', Validators.required],
        https: [''],
        allowedOrigins: [[]]
      }),
      oauth2: this.fb.group({
        skipApprovalScreen: [false],
        alwaysShowLoginScreen: [false],
        passwordConnector: ['']
      }),
      enablePasswordDB: [false],
      staticClients: this.fb.array([]),
      connectors: this.fb.array([])
    });

    // Add default Ganje client
    this.addStaticClient();
    const ganjeClient = this.staticClientsArray.at(0);
    ganjeClient.patchValue({
      id: 'ganje-admin-portal',
      name: 'Ganje Admin Portal',
      redirectURIs: 'http://localhost:4200/auth/callback\nhttps://ganje.example.com/auth/callback',
      public: false
    });
  }

  get staticClientsArray(): FormArray {
    return this.configForm.get('staticClients') as FormArray;
  }

  get connectorsArray(): FormArray {
    return this.configForm.get('connectors') as FormArray;
  }

  private async loadConfiguration() {
    const loadingKey = 'dex-config-load';
    this.loading = true;
    this.loadingService.setLoading(loadingKey, true);
    
    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      // Mock configuration data
      const mockConfig = {
        issuer: 'https://dex.example.com',
        storage: {
          type: 'postgres',
          config: {
            host: 'localhost',
            port: 5432,
            database: 'dex',
            user: 'dex',
            password: 'password',
            ssl: false
          }
        },
        web: {
          http: '0.0.0.0:5556',
          https: '',
          allowedOrigins: ['*']
        },
        oauth2: {
          skipApprovalScreen: false,
          alwaysShowLoginScreen: false,
          passwordConnector: 'local'
        },
        enablePasswordDB: true,
        staticClients: [
          {
            id: 'ganje-admin',
            redirectURIs: ['http://localhost:4200/callback'],
            name: 'Ganje Admin Portal',
            secret: 'ZXhhbXBsZS1hcHAtc2VjcmV0'
          }
        ],
        connectors: [
          {
            type: 'ldap',
            id: 'ldap',
            name: 'LDAP',
            config: {
              host: 'ldap.example.com:389',
              bindDN: 'cn=admin,dc=example,dc=com',
              bindPW: 'admin',
              userSearch: {
                baseDN: 'ou=People,dc=example,dc=com',
                filter: '(objectClass=person)',
                username: 'uid',
                idAttr: 'uid',
                emailAttr: 'mail',
                nameAttr: 'cn'
              }
            }
          }
        ]
      };
      
      this.populateForm(mockConfig);
    } catch (error) {
      this.errorHandler.handleError(error, 'Loading Dex configuration');
    } finally {
      this.loading = false;
      this.loadingService.setLoading(loadingKey, false);
    }
  }

  private populateForm(config: DexConfig) {
    this.configForm.patchValue(config);
    this.storageType = config.storage.type;

    // Process static clients
    config.staticClients.forEach((client, index) => {
      if (index < this.staticClientsArray.length) {
        this.staticClientsArray.at(index).patchValue(client);
      } else {
        this.addStaticClient();
        this.staticClientsArray.at(index).patchValue(client);
      }
    });

    // Process connectors
    config.connectors.forEach((connector, index) => {
      if (index < this.connectorsArray.length) {
        this.connectorsArray.at(index).patchValue(connector);
      } else {
        this.addConnector();
        this.connectorsArray.at(index).patchValue(connector);
      }
    });
  }

  onStorageTypeChange(type: string) {
    this.storageType = type;
    const storageConfig = this.configForm.get('storage.config') as FormGroup;
    
    // Clear existing config
    Object.keys(storageConfig.controls).forEach(key => {
      storageConfig.removeControl(key);
    });

    // Add type-specific controls
    switch (type) {
      case 'sqlite3':
        storageConfig.addControl('file', this.fb.control('/var/dex/dex.db'));
        break;
      case 'postgres':
        storageConfig.addControl('host', this.fb.control('localhost'));
        storageConfig.addControl('port', this.fb.control(5432));
        storageConfig.addControl('database', this.fb.control('dex'));
        storageConfig.addControl('user', this.fb.control('dex'));
        storageConfig.addControl('password', this.fb.control(''));
        storageConfig.addControl('ssl', this.fb.control(false));
        break;
      case 'mysql':
        storageConfig.addControl('host', this.fb.control('localhost'));
        storageConfig.addControl('port', this.fb.control(3306));
        storageConfig.addControl('database', this.fb.control('dex'));
        storageConfig.addControl('user', this.fb.control('dex'));
        storageConfig.addControl('password', this.fb.control(''));
        storageConfig.addControl('ssl', this.fb.control(false));
        break;
    }
  }

  getStorageConfigControl(controlName: string): FormControl {
    return this.configForm.get(`storage.config.${controlName}`) as FormControl;
  }

  addStaticClient() {
    const clientGroup = this.fb.group({
      id: ['', Validators.required],
      name: ['', Validators.required],
      secret: [''],
      redirectURIs: ['', Validators.required],
      public: [false],
      trustedPeers: [[]]
    });
    
    this.staticClientsArray.push(clientGroup);
  }

  removeStaticClient(index: number) {
    this.staticClientsArray.removeAt(index);
  }

  addConnector() {
    const connectorGroup = this.fb.group({
      type: ['', Validators.required],
      id: ['', Validators.required],
      name: ['', Validators.required],
      config: ['{}', Validators.required]
    });
    
    this.connectorsArray.push(connectorGroup);
  }

  removeConnector(index: number) {
    this.connectorsArray.removeAt(index);
  }

  async saveConfiguration() {
    if (this.configForm.invalid) {
      this.markFormGroupTouched(this.configForm);
      return;
    }

    const loadingKey = 'dex-config-save';
    this.saving = true;
    this.loadingService.setLoading(loadingKey, true);
    
    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 1500));
      
      const formValue = this.configForm.value;
      console.log('Saving configuration:', formValue);
      
      this.errorHandler.handleSuccess('Configuration saved successfully!');
    } catch (error) {
      this.errorHandler.handleError(error, 'Saving Dex configuration');
    } finally {
      this.saving = false;
      this.loadingService.setLoading(loadingKey, false);
    }
  }

  private prepareConfigForSave(): DexConfig {
    const formValue = this.configForm.value;
    
    // Process static clients
    const staticClients = formValue.staticClients.map((client: any) => ({
      ...client,
      redirectURIs: client.redirectURIs.split('\n').filter((uri: string) => uri.trim())
    }));

    // Process connectors
    const connectors = formValue.connectors.map((connector: any) => ({
      ...connector,
      config: JSON.parse(connector.config || '{}')
    }));

    return {
      ...formValue,
      staticClients,
      connectors
    };
  }

  resetToDefaults() {
    const confirmed = confirm('Are you sure you want to reset to default configuration? This will lose all current settings.');
    if (confirmed) {
      this.initializeForm();
      this.errorHandler.handleInfo('Configuration reset to defaults');
    }
  }

  private markFormGroupTouched(formGroup: FormGroup) {
    Object.keys(formGroup.controls).forEach(key => {
      const control = formGroup.get(key);
      if (control instanceof FormGroup) {
        this.markFormGroupTouched(control);
      } else if (control instanceof FormArray) {
        control.controls.forEach(arrayControl => {
          if (arrayControl instanceof FormGroup) {
            this.markFormGroupTouched(arrayControl);
          } else {
            arrayControl.markAsTouched();
          }
        });
      } else {
        control?.markAsTouched();
      }
    });
  }
}
