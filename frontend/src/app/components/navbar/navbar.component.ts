import { Component, EventEmitter, Output } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatListModule } from '@angular/material/list';
import { MatToolbarModule } from '@angular/material/toolbar';
import { ThemeService } from '@core/services/theme.service';

@Component({
  selector: 'app-navbar',
  imports: [MatListModule,
    MatToolbarModule,
    MatButtonModule,
    MatIconModule],
  templateUrl: './navbar.component.html',
  styleUrl: './navbar.component.scss',
})
export class NavbarComponent {
  @Output() menuToggle = new EventEmitter<void>();

  constructor(public themeService: ThemeService) {}
}
