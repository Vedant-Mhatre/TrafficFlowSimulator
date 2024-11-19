#include <iostream>
#include <vector>
#include <chrono>
#include <thread>

// Constants for grid dimensions and simulation settings
const int GRID_WIDTH = 20;
const int GRID_HEIGHT = 10;
const int SIMULATION_STEPS = 50;

void clearScreen()
{
#ifdef _WIN32
    system("cls");
#else
    system("clear");
#endif
}

int main()
{
    // Simulation loop
    for (int step = 0; step < SIMULATION_STEPS; ++step)
    {
        clearScreen();
        std::cout << "Step: " << step + 1 << std::endl;

        std::this_thread::sleep_for(std::chrono::milliseconds(500));
    }

    std::cout << "Simulation Complete!" << std::endl;
    return 0;
}
