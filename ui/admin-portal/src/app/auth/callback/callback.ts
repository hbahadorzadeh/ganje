import { Component } from '@angular/core';
import { Router, ActivatedRoute } from '@angular/router';
import { AuthService } from '../../core/services/auth';

@Component({
  selector: 'app-callback',
  imports: [],
  templateUrl: './callback.html',
  styleUrl: './callback.scss'
})
export class Callback {
  constructor(private auth: AuthService, private router: Router, private route: ActivatedRoute) {
    this.finishLogin();
  }

  async finishLogin() {
    await this.auth.processCallback();
    const returnUrl = this.route.snapshot.queryParamMap.get('returnUrl') || '/repositories';
    this.router.navigateByUrl(returnUrl);
  }
}
