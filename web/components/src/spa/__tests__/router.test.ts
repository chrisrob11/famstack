/**
 * Find router navigation issue
 */

import { SPARouter, routerConfig } from '../router';

// Mock DOM with app container
document.body.innerHTML = '<div id="app"></div>';

describe('Router Issue', () => {
  it('can navigate to tasks', async () => {
    const router = new SPARouter(routerConfig);

    // Mock the loadComponent method to prevent actual component loading
    const loadComponentSpy = jest.spyOn(router as any, 'loadComponent').mockResolvedValue(undefined);

    await router.navigate('/tasks');

    expect(loadComponentSpy).toHaveBeenCalled();
  });
});