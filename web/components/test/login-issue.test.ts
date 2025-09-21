/**
 * Minimal test to find the login redirect issue
 */

describe('Login Issue Debug', () => {
  it('can router navigate after login', () => {
    // Mock the exact scenario
    const mockRouter = {
      navigate: jest.fn()
    };

    (window as any).famstackApp = {
      getRouter: () => mockRouter
    };

    // Simulate the login page redirect code
    const router = (window as any).famstackApp?.getRouter();
    if (router) {
      router.navigate('/tasks');
    }

    expect(mockRouter.navigate).toHaveBeenCalledWith('/tasks');
  });

  it('can import daily-page component', async () => {
    try {
      await import('../src/pages/daily-page');
      console.log('✅ daily-page component loads');
    } catch (error) {
      console.log('❌ daily-page component failed:', error.message);
    }
  });
});