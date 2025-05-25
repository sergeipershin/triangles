"""
This module provides classes and functions to generate and visualize
patterns, which are figures made of identical equilateral triangles
connected edge-to-edge on a plane.

The script defines a triangular grid system using three axes at 120 degrees
to each other. Triangles are identified by coordinates (x, y, z)
representing the three grid lines forming their sides. The sum x+y+z
is +1 or -1, indicating the triangle's orientation.

The core functionality includes:
- Representing individual triangles (`Triangle` class) and their transformations.
- Representing patterns of triangles (`Pattern` class), including methods
  for transformations, normalization, and symmetry-aware equality checking.
- Visualization of patterns as images using PIL.
"""

from PIL import Image, ImageDraw
import math
from pathlib import Path


class Axes:
    """Represents the three axes in the triangular coordinate system."""

    def __init__(self):
        self._axes = ['x', 'y', 'z']

    def __iter__(self):
        return iter(self._axes)

    def __len__(self):
        return len(self._axes)

    def __contains__(self, axis):
        return axis in self._axes

    def __getitem__(self, item):
        return self._axes[item]


class Triangle:
    """Represents a single equilateral triangle in the triangular grid."""

    def __init__(self, x, y, z):
        self.x = x
        self.y = y
        self.z = z

    def __key(self):
        return self.x, self.y, self.z

    def __hash__(self):
        return hash(self.__key())

    def __eq__(self, other):
        return (self.x == other.x) and (self.y == other.y) and (self.z == other.z)

    def __str__(self):
        return f'Triangle({self.x}, {self.y}, {self.z})'

    def get_coord(self, axis):
        """Returns the coordinate value for the specified axis."""
        match axis:
            case 'x':
                return self.x
            case 'y':
                return self.y
            case 'z':
                return self.z
            case _:
                return None

    def get_neighbour_coords(self, axis):
        """Returns coordinates of the neighboring triangle along the specified axis."""
        look = self.x + self.y + self.z
        match axis:
            case 'x':
                return self.x, self.y - look, self.z - look
            case 'y':
                return self.x - look, self.y, self.z - look
            case 'z':
                return self.x - look, self.y - look, self.z
            case _:
                return None

    def get_neighbour(self, axis):
        """Returns the neighboring triangle along the specified axis."""
        x, y, z = self.get_neighbour_coords(axis)
        return Triangle(x, y, z)

    def get_rotated_coords(self, angle):
        """Returns coordinates after rotation by the specified angle (multiple of 60 degrees)."""
        match angle:
            case 60:
                return -self.y, -self.z, -self.x
            case 120:
                return self.z, self.x, self.y
            case 180:
                return -self.x, -self.y, -self.z
            case 240:
                return self.y, self.z, self.x
            case 300:
                return -self.z, -self.x, -self.y
            case _:
                return None

    def get_rotated(self, angle):
        """Returns a new triangle rotated by the specified angle."""
        x, y, z = self.get_rotated_coords(angle)
        return Triangle(x, y, z)

    def get_reflected_coords(self, axis):
        """Returns coordinates after reflection over the specified axis."""
        match axis:
            case 'x':
                return -self.x, -self.z, -self.y
            case 'y':
                return -self.z, -self.y, -self.x
            case 'z':
                return -self.y, -self.x, -self.z
            case _:
                return None

    def get_reflected(self, axis):
        """Returns a new triangle reflected over the specified axis."""
        x, y, z = self.get_reflected_coords(axis)
        return Triangle(x, y, z)

    def get_shifted_coords(self, shift, axis):
        """Returns coordinates after shifting along the specified axis."""
        match axis:
            case 'x':
                return self.x, self.y + shift, self.z - shift
            case 'y':
                return self.x - shift, self.y, self.z + shift
            case 'z':
                return self.x + shift, self.y - shift, self.z
            case _:
                return None

    def get_shifted(self, shift, axis):
        """Returns a new triangle shifted along the specified axis."""
        x, y, z = self.get_shifted_coords(shift, axis)
        return Triangle(x, y, z)

    def get_copy(self):
        """Returns a copy of this triangle."""
        return Triangle(self.x, self.y, self.z)

    def get_cartesian_coords(self, axis='all'):
        """Converts triangular coordinates to Cartesian coordinates for drawing."""
        tg30 = math.tan(math.pi / 6.0)
        match axis:
            case 'x':
                x1 = self.x / (2.0 * tg30)
                y1 = self.x / 2.0 + self.y
                x2 = x1
                y2 = -self.x / 2.0 - self.z
                return (x1, y1), (x2, y2)
            case 'y':
                x1 = self.x / (2.0 * tg30)
                y1 = self.x / 2.0 + self.y
                x2 = -(self.z + self.y) / (2.0 * tg30)
                y2 = x2 * tg30 + self.y
                return (x1, y1), (x2, y2)
            case 'z':
                x1 = self.x / (2.0 * tg30)
                y1 = -self.x / 2.0 - self.z
                x2 = -(self.z + self.y) / (2.0 * tg30)
                y2 = -x2 * tg30 - self.z
                return (x1, y1), (x2, y2)
            case 'all':
                x1 = self.x / (2.0 * tg30)
                y1 = self.x / 2.0 + self.y
                x2 = x1
                y2 = -self.x / 2.0 - self.z
                x3 = -(self.z + self.y) / (2.0 * tg30)
                y3 = x3 * tg30 + self.y
                return (x1, y1), (x2, y2), (x3, y3)
            case _:
                return None


class Pattern:
    """Represents a collection of triangles forming a pattern."""

    def __init__(self):
        self._triangles = []

    def __iter__(self):
        return iter(self._triangles)

    def __len__(self):
        return len(self._triangles)

    def __contains__(self, item):
        return item in self._triangles

    def __eq__(self, other):
        """Checks if two patterns are equivalent under rotation/reflection."""
        found_equal = False
        aligned_self = self.get_aligned('z')
        rotated_other = other
        for _ in range(6):
            rotated_other = rotated_other.get_rotated(60)
            for i in range(2):
                if i == 0:
                    aligned_other = rotated_other.get_aligned('z')
                else:
                    aligned_other = rotated_other.get_reflected('z').get_aligned('z')
                are_equal = True
                for triangle in aligned_self:
                    if triangle not in aligned_other:
                        are_equal = False
                        break
                if are_equal:
                    for triangle in aligned_other:
                        if triangle not in aligned_self:
                            are_equal = False
                            break
                if are_equal:
                    found_equal = True
                    break
            if found_equal:
                break
        return found_equal

    def __str__(self):
        return f'Triangle pattern: {', '.join([str(triangle) for triangle in self])}'

    def add_triangle(self, triangle):
        """Adds a triangle to the pattern."""
        self._triangles.append(triangle)

    def get_min_coord(self, axis):
        """Returns the minimum coordinate value along the specified axis."""
        curr_min = None
        for triangle in self._triangles:
            if curr_min is None:
                curr_min = triangle.get_coord(axis)
            elif curr_min > triangle.get_coord(axis):
                curr_min = triangle.get_coord(axis)
        return curr_min

    def get_max_coord(self, axis):
        """Returns the maximum coordinate value along the specified axis."""
        curr_max = None
        for triangle in self._triangles:
            if curr_max is None:
                curr_max = triangle.get_coord(axis)
            elif curr_max < triangle.get_coord(axis):
                curr_max = triangle.get_coord(axis)
        return curr_max

    def get_shifted(self, shift, axis):
        """Returns a new pattern shifted along the specified axis."""
        shifted = Pattern()
        for triangle in self:
            shifted.add_triangle(triangle.get_shifted(shift, axis))
        return shifted

    def get_rotated(self, angle):
        """Returns a new pattern rotated by the specified angle."""
        rotated = Pattern()
        for triangle in self:
            rotated.add_triangle(triangle.get_rotated(angle))
        return rotated

    def get_reflected(self, axis):
        """Returns a new pattern reflected over the specified axis."""
        reflected = Pattern()
        for triangle in self:
            reflected.add_triangle(triangle.get_reflected(axis))
        return reflected

    def get_aligned(self, free_axis):
        """Aligns the pattern along the specified free axis."""
        match free_axis:
            case 'x':
                max_coord = self.get_max_coord('y')
                aligned = self.get_shifted(max_coord, 'z')
                min_coord = aligned.get_min_coord('z')
                aligned = aligned.get_shifted(-min_coord, 'y')
            case 'y':
                max_coord = self.get_max_coord('z')
                aligned = self.get_shifted(max_coord, 'x')
                min_coord = aligned.get_min_coord('x')
                aligned = aligned.get_shifted(-min_coord, 'z')
            case 'z':
                max_coord = self.get_max_coord('x')
                aligned = self.get_shifted(max_coord, 'y')
                min_coord = aligned.get_min_coord('y')
                aligned = aligned.get_shifted(-min_coord, 'x')
            case _:
                aligned = None
        return aligned

    def get_centered(self):
        """Centers the pattern around the origin."""
        min_coord = self.get_min_coord('x')
        max_coord = self.get_max_coord('x')
        mean_coord = int((min_coord + max_coord) / 2)
        centered = self.get_shifted(mean_coord, 'y')
        min_coord = centered.get_min_coord('y')
        max_coord = centered.get_max_coord('y')
        mean_coord = int((min_coord + max_coord) / 2)
        centered = centered.get_shifted(-mean_coord, 'x')
        return centered

    def get_copy(self):
        """Returns a deep copy of the pattern."""
        pattern_copy = Pattern()
        for triangle in self:
            pattern_copy.add_triangle(triangle.get_copy())
        return pattern_copy

    def create_image(self, show_axes=True, show_grid=True):
        """Creates an image representation of the pattern."""
        tg30 = math.tan(math.pi / 6.0)
        axes = Axes()

        # Calculate the bounding radius for the image
        lines = []
        radius = 0
        for triangle in self:
            for axis in axes:
                line = triangle.get_cartesian_coords(axis)
                if abs(line[0][0]) > radius:
                    radius = abs(line[0][0])
                if abs(line[1][0]) > radius:
                    radius = abs(line[1][0])
                if abs(line[0][1]) > radius:
                    radius = abs(line[0][1])
                if abs(line[1][1]) > radius:
                    radius = abs(line[1][1])
                neighbour = triangle.get_neighbour(axis)
                if neighbour in self:
                    lines.append([line, 'normal'])
                else:
                    lines.append([line, 'bold'])

        # Set up image dimensions
        x_min = -int(radius) - 1
        y_min = x_min
        x_max = -x_min
        y_max = x_max
        scale = 200

        img_width = (x_max - x_min) * scale + 20
        img_height = (y_max - y_min) * scale + 20

        img = Image.new('RGB', (img_width, img_height), color='white')
        draw = ImageDraw.Draw(img)

        def to_real(_x, _y):
            """Converts coordinates to image pixel coordinates."""
            return _x * scale + img_width / 2, img_height / 2 - _y * scale

        # Draw grid lines if enabled
        if show_grid:
            for x in range(int(x_min * 2.0 * tg30), int(x_max * 2.0 * tg30) + 1):
                x1, y1 = to_real(x / (2.0 * tg30), y_min)
                x2, y2 = to_real(x / (2.0 * tg30), y_max)
                draw.line([x1, y1, x2, y2], fill='lightgray', width=1)
            for y in range(y_max - int(x_min * tg30), y_min - int(x_max * tg30), -1):
                x1 = x_min
                y1 = y + x_min * tg30
                x2 = x_max
                y2 = y1 + (x_max - x_min) * tg30
                if y1 < y_min:
                    x1 = x_min + (y_min - y1) / tg30
                    y1 = y_min
                if y2 > y_max:
                    x2 = x_max - (y2 - y_max) / tg30
                    y2 = y_max
                x3 = x_max - x1 + x_min
                x4 = x_max - x2 + x_min
                x1, y1 = to_real(x1, y1)
                x2, y2 = to_real(x2, y2)
                x3, _ = to_real(x3, y1)
                x4, _ = to_real(x4, y2)
                draw.line([x1, y1, x2, y2], fill='lightgray', width=1)
                draw.line([x3, y1, x4, y2], fill='lightgray', width=1)

        # Draw axes if enabled
        if show_axes:
            x0 = 0
            y0 = 0
            x1 = 0
            y1 = y_max
            x2 = x_min
            y2 = x_min * tg30
            x3 = x_max
            y3 = -x_max * tg30
            x0, y0 = to_real(x0, y0)
            x1, y1 = to_real(x1, y1)
            x2, y2 = to_real(x2, y2)
            x3, y3 = to_real(x3, y3)
            draw.line([x0, y0, x1, y1], fill='gray', width=1)
            draw.line([x0, y0, x2, y2], fill='gray', width=1)
            draw.line([x0, y0, x3, y3], fill='gray', width=1)

        # Draw all triangle edges
        for line in lines:
            x1, y1 = to_real(line[0][0][0], line[0][0][1])
            x2, y2 = to_real(line[0][1][0], line[0][1][1])
            if line[1] == 'bold':
                draw.line([x1, y1, x2, y2], fill='black', width=4)
            else:
                draw.line([x1, y1, x2, y2], fill='black', width=1)

        return img

    def dump_to_file(self, file_path):
        """Saves the pattern image to a file."""
        img = self.create_image()
        img.save(file_path)


def generate_patterns(patterns, triangles_to_add, sketch=None):
    """
    Recursively generates all unique patterns with the specified number of triangles.
    Patterns that are rotations or reflections of each other are considered identical.
    """
    axes = Axes()
    if sketch is None:
        # Initialize with a single triangle
        sketch = Pattern()
        triangle = Triangle(0, 1, 0)
        sketch.add_triangle(triangle)
        if triangles_to_add > 1:
            generate_patterns(patterns, triangles_to_add - 1, sketch)
        else:
            patterns.append(sketch)
        return

    # Try adding neighbors to each existing triangle
    for triangle in sketch:
        for axis in axes:
            neighbour = triangle.get_neighbour(axis)
            if neighbour in sketch:
                continue
            new_sketch = sketch.get_copy()
            new_sketch.add_triangle(neighbour)
            if triangles_to_add > 1:
                generate_patterns(patterns, triangles_to_add - 1, new_sketch)
            else:
                # Check if this is a new unique pattern
                found_new_pattern = True
                for pattern in patterns:
                    if pattern == new_sketch:
                        found_new_pattern = False
                        break
                if found_new_pattern:
                    patterns.append(new_sketch.get_centered())
                    return


def main():
    """Main function to generate and save patterns."""
    patterns = []
    num_triangles = int(input('Enter number of triangles: '))
    generate_patterns(patterns, num_triangles)

    # Create output directory and save images
    work_path = Path('.')
    output_path = work_path / str(num_triangles)
    output_path.mkdir(exist_ok=True)
    for i, pattern in enumerate(patterns):
        pattern.dump_to_file(output_path / f'{num_triangles}_{i}.png')


if __name__ == '__main__':
    main()
