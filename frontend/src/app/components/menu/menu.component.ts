import { Component, Inject, PLATFORM_ID, signal } from '@angular/core';
import { MatIconModule } from '@angular/material/icon';
import { MatButtonModule } from '@angular/material/button';
import { BreakpointObserver, Breakpoints } from '@angular/cdk/layout';
import { isPlatformBrowser } from '@angular/common';
import { MENU_ITEMS, MenuItem } from '../../app.routes';
import { MatListModule } from '@angular/material/list';
import { RouterModule } from '@angular/router';

@Component({
  selector: 'app-menu',
  imports: [MatListModule, MatButtonModule, RouterModule, MatIconModule],
  templateUrl: './menu.component.html',
  styleUrl: './menu.component.scss',
})
export class MenuComponent {
  isMobile = signal(false);
  menuItems: MenuItem[] = MENU_ITEMS;

  constructor(
    @Inject(PLATFORM_ID) private platformId: Object,
    private bp: BreakpointObserver
  ) { }

  ngOnInit() {
    if (isPlatformBrowser(this.platformId)) {
      this.isMobile.set(this.bp.isMatched(Breakpoints.Handset));

      this.bp.observe([Breakpoints.Handset]).subscribe((result) => {
        this.isMobile.set(result.matches);
      });
    }
  }
}
